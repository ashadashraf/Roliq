[CmdletBinding()]
param(
  [string]$ApiBaseUrl = "http://localhost:8080",
  [string]$EnvironmentFile = ".env"
)

$ErrorActionPreference = "Stop"

function Assert-True {
  param([bool]$Condition, [string]$Message)
  if (-not $Condition) {
    throw "Acceptance assertion failed: $Message"
  }
}

function Read-DotEnv {
  param([string]$Path)
  $values = @{}
  Get-Content -LiteralPath $Path | ForEach-Object {
    if ($_ -match '^\s*([^#][^=]*)=(.*)$') {
      $values[$matches[1].Trim()] = $matches[2].Trim()
    }
  }
  return $values
}

function Invoke-Clerk {
  param(
    [string]$Method,
    [string]$Path,
    [hashtable]$Headers,
    [object]$Body
  )
  $parameters = @{
    Method  = $Method
    Uri     = "https://api.clerk.com/v1$Path"
    Headers = $Headers
  }
  if ($null -ne $Body) {
    $parameters.Body = $Body | ConvertTo-Json -Depth 10
  }
  return Invoke-RestMethod @parameters
}

function New-TestIdentity {
  param([string]$Label, [hashtable]$ClerkHeaders)
  $stamp = [DateTimeOffset]::UtcNow.ToUnixTimeMilliseconds()
  $user = Invoke-Clerk -Method Post -Path "/users" -Headers $ClerkHeaders -Body @{
    email_address = @("roliq-phase1-$Label-$stamp+clerk_test@example.com")
    first_name    = "Phase"
    last_name     = $Label
    password      = "Roliq!$stamp-Aa"
  }
  $session = Invoke-Clerk -Method Post -Path "/sessions" -Headers $ClerkHeaders -Body @{ user_id = $user.id }
  $token = Invoke-Clerk -Method Post -Path "/sessions/$($session.id)/tokens/roliq-api" -Headers $ClerkHeaders -Body @{}
  return [pscustomobject]@{ UserId = $user.id; Token = $token.jwt }
}

function Invoke-Api {
  param(
    [string]$Method,
    [string]$Path,
    [string]$Token,
    [object]$Body,
    [hashtable]$AdditionalHeaders = @{}
  )
  $headers = @{ Authorization = "Bearer $Token" }
  foreach ($key in $AdditionalHeaders.Keys) {
    $headers[$key] = $AdditionalHeaders[$key]
  }
  $parameters = @{
    Method      = $Method
    Uri         = "$ApiBaseUrl$Path"
    Headers     = $headers
    ContentType = "application/json"
  }
  if ($null -ne $Body) {
    $parameters.Body = $Body | ConvertTo-Json -Depth 10
  }
  return Invoke-RestMethod @parameters
}

function Get-ApiFailureStatus {
  param([string]$Method, [string]$Path, [string]$Token, [hashtable]$AdditionalHeaders = @{})
  try {
    Invoke-Api -Method $Method -Path $Path -Token $Token -Body $null -AdditionalHeaders $AdditionalHeaders | Out-Null
    return 200
  }
  catch {
    return [int]$_.Exception.Response.StatusCode
  }
}

function Get-Sha256 {
  param([byte[]]$Bytes)
  $algorithm = [Security.Cryptography.SHA256]::Create()
  try {
    return -join ($algorithm.ComputeHash($Bytes) | ForEach-Object { $_.ToString("x2") })
  }
  finally {
    $algorithm.Dispose()
  }
}

function New-DocxFixture {
  Add-Type -AssemblyName System.IO.Compression
  $stream = [IO.MemoryStream]::new()
  $archive = [IO.Compression.ZipArchive]::new($stream, [IO.Compression.ZipArchiveMode]::Create, $true)
  try {
    $entries = @{
      "[Content_Types].xml" = '<?xml version="1.0" encoding="UTF-8"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="xml" ContentType="application/xml"/></Types>'
      "word/document.xml"   = '<?xml version="1.0" encoding="UTF-8"?><w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body><w:p><w:r><w:t>Roliq Phase 1 acceptance fixture</w:t></w:r></w:p></w:body></w:document>'
    }
    foreach ($name in $entries.Keys) {
      $entry = $archive.CreateEntry($name)
      $writer = [IO.StreamWriter]::new($entry.Open(), [Text.UTF8Encoding]::new($false))
      try { $writer.Write($entries[$name]) } finally { $writer.Dispose() }
    }
  }
  finally {
    $archive.Dispose()
  }
  $bytes = $stream.ToArray()
  $stream.Dispose()
  return $bytes
}

function New-UploadIntent {
  param(
    [pscustomobject]$Identity,
    [string]$FileName,
    [string]$ContentType,
    [byte[]]$Bytes,
    [string]$Checksum,
    [string]$IdempotencyKey
  )
  return Invoke-Api -Method Post -Path "/v1/resume-uploads" -Token $Identity.Token -Body @{
    fileName       = $FileName
    contentType    = $ContentType
    sizeBytes      = $Bytes.Length
    checksumSha256 = $Checksum
  } -AdditionalHeaders @{ "Idempotency-Key" = $IdempotencyKey }
}

function Send-Upload {
  param([object]$Intent, [string]$ContentType, [byte[]]$Bytes, [string]$Checksum)
  Invoke-WebRequest -UseBasicParsing -Method Put -Uri $Intent.uploadUrl -ContentType $ContentType -Headers @{
    "x-amz-meta-sha256" = $Checksum
  } -Body $Bytes | Out-Null
}

function Wait-ResumeStatus {
  param([pscustomobject]$Identity, [string]$ResumeId, [string[]]$Expected, [int]$TimeoutSeconds = 45)
  $deadline = (Get-Date).AddSeconds($TimeoutSeconds)
  do {
    $list = Invoke-Api -Method Get -Path "/v1/resumes" -Token $Identity.Token -Body $null
    $resume = @($list.items | Where-Object { $_.id -eq $ResumeId })[0]
    if ($null -ne $resume -and $Expected -contains $resume.status) {
      return $resume
    }
    Start-Sleep -Seconds 2
  } while ((Get-Date) -lt $deadline)
  throw "Resume $ResumeId did not reach one of: $($Expected -join ', ')"
}

function Invoke-DockerSql {
  param([string]$Sql)
  & docker compose --env-file $EnvironmentFile exec -T postgres psql -v ON_ERROR_STOP=1 -U postgres -d roliq -c $Sql | Out-Null
  Assert-True ($LASTEXITCODE -eq 0) "PostgreSQL acceptance command failed"
}

function Invoke-DockerSqlScalar {
  param([string]$Sql, [switch]$ApplicationRole)
  if ($ApplicationRole) {
    $output = & docker compose --env-file $EnvironmentFile exec -T -e PGPASSWORD=roliq_app postgres psql -h 127.0.0.1 -U roliq_app -d roliq -Atqc $Sql
  }
  else {
    $output = & docker compose --env-file $EnvironmentFile exec -T postgres psql -U postgres -d roliq -Atqc $Sql
  }
  Assert-True ($LASTEXITCODE -eq 0) "PostgreSQL scalar acceptance command failed"
  return (($output | Out-String).Trim())
}

function Wait-LocalStackResources {
  param([int]$TimeoutSeconds = 45)
  $deadline = (Get-Date).AddSeconds($TimeoutSeconds)
  do {
    & docker compose --env-file $EnvironmentFile exec -T localstack awslocal sqs get-queue-url --queue-name roliq-domain-events | Out-Null
    if ($LASTEXITCODE -eq 0) {
      & docker compose --env-file $EnvironmentFile exec -T localstack awslocal s3api head-bucket --bucket roliq-resumes-local | Out-Null
      if ($LASTEXITCODE -eq 0) { return }
    }
    Start-Sleep -Seconds 2
  } while ((Get-Date) -lt $deadline)
  throw "LocalStack S3 and SQS resources did not become ready"
}

$environment = Read-DotEnv -Path $EnvironmentFile
Assert-True ($environment.ContainsKey("CLERK_SECRET_KEY")) "CLERK_SECRET_KEY is missing"
$clerkHeaders = @{ Authorization = "Bearer $($environment['CLERK_SECRET_KEY'])"; "Content-Type" = "application/json" }
$templates = @(Invoke-Clerk -Method Get -Path "/jwt_templates?limit=100" -Headers $clerkHeaders -Body $null)
Assert-True (@($templates | Where-Object { $_.name -eq "roliq-api" }).Count -eq 1) "Clerk JWT template roliq-api is missing"

$identities = @()
$organizations = @()
$localstackStopped = $false
try {
  $manualIdentity = New-TestIdentity -Label "Manual" -ClerkHeaders $clerkHeaders
  $resumeIdentity = New-TestIdentity -Label "Resume" -ClerkHeaders $clerkHeaders
  $identities = @($manualIdentity, $resumeIdentity)

  $manualSession = Invoke-Api -Method Post -Path "/v1/session/bootstrap" -Token $manualIdentity.Token -Body $null
  $manualSessionAgain = Invoke-Api -Method Post -Path "/v1/session/bootstrap" -Token $manualIdentity.Token -Body $null
  $resumeSession = Invoke-Api -Method Post -Path "/v1/session/bootstrap" -Token $resumeIdentity.Token -Body $null
  $organizations = @($manualSession.organization.id, $resumeSession.organization.id)
  Assert-True ($manualSession.user.id -eq $manualSessionAgain.user.id) "session bootstrap is not idempotent"
  Assert-True ($manualSession.organization.id -eq $manualSessionAgain.organization.id) "personal organization bootstrap is not idempotent"

  $crossTenantStatus = Get-ApiFailureStatus -Method Get -Path "/v1/me" -Token $manualIdentity.Token -AdditionalHeaders @{
    "X-Organization-ID" = $resumeSession.organization.id
  }
  Assert-True ($crossTenantStatus -eq 403) "cross-tenant organization selection was not denied"
  $unauthorizedStatus = Get-ApiFailureStatus -Method Get -Path "/v1/me" -Token "invalid-token"
  Assert-True ($unauthorizedStatus -eq 401) "invalid bearer token was not rejected"

  Invoke-Api -Method Patch -Path "/v1/onboarding" -Token $manualIdentity.Token -Body @{
    currentStep = 2; status = "in_progress"; profileMethod = "manual"
  } | Out-Null
  $profile = Invoke-Api -Method Put -Path "/v1/career-profile" -Token $manualIdentity.Token -Body @{
    headline        = "Platform engineer"
    summary         = "Phase 1 persisted-profile acceptance record."
    countryCode     = "IN"
    timeZone        = "Asia/Kolkata"
    city            = "Bengaluru"
    yearsExperience = 5
    skills          = @("Go", "PostgreSQL", "TypeScript")
    experiences     = @(@{ company = "Acceptance Systems"; title = "Engineer"; location = "Remote"; startDate = "2021-01-01"; isCurrent = $true; description = "Acceptance-only record." })
    education       = @(@{ institution = "Acceptance University"; degree = "BSc"; fieldOfStudy = "Computer Science" })
  }
  Invoke-Api -Method Patch -Path "/v1/onboarding" -Token $manualIdentity.Token -Body @{
    currentStep = 4; status = "completed"; profileMethod = "manual"
  } | Out-Null
  $reloadedProfile = Invoke-Api -Method Get -Path "/v1/career-profile" -Token $manualIdentity.Token -Body $null
  $dashboard = Invoke-Api -Method Get -Path "/v1/dashboard" -Token $manualIdentity.Token -Body $null
  Assert-True ($reloadedProfile.headline -eq $profile.headline) "career profile did not persist across requests"
  Assert-True ($dashboard.profileCompletion -eq 100) "persisted dashboard profile completion is incorrect"
  Assert-True ($dashboard.onboarding.status -eq "completed") "manual onboarding did not persist"

  $ownProfileCount = Invoke-DockerSqlScalar -ApplicationRole -Sql "SET app.user_id='$($manualSession.user.id)'; SET app.organization_id='$($manualSession.organization.id)'; SELECT count(*) FROM career_profiles WHERE headline='Platform engineer';"
  $crossTenantProfileCount = Invoke-DockerSqlScalar -ApplicationRole -Sql "SET app.user_id='$($manualSession.user.id)'; SET app.organization_id='$($resumeSession.organization.id)'; SELECT count(*) FROM career_profiles WHERE headline='Platform engineer';"
  Assert-True ($ownProfileCount -eq "1") "application role could not read its tenant profile"
  Assert-True ($crossTenantProfileCount -eq "0") "RLS exposed a profile across organization boundaries"

  $pdf = [Text.Encoding]::ASCII.GetBytes("%PDF-1.4`n% Roliq Phase 1 acceptance`n%%EOF`n")
  $pdfChecksum = Get-Sha256 -Bytes $pdf
  $idempotencyKey = "phase1-acceptance-$([Guid]::NewGuid())"
  $intent = New-UploadIntent -Identity $resumeIdentity -FileName "acceptance.pdf" -ContentType "application/pdf" -Bytes $pdf -Checksum $pdfChecksum -IdempotencyKey $idempotencyKey
  $replayedIntent = New-UploadIntent -Identity $resumeIdentity -FileName "acceptance.pdf" -ContentType "application/pdf" -Bytes $pdf -Checksum $pdfChecksum -IdempotencyKey $idempotencyKey
  Assert-True ($intent.uploadId -eq $replayedIntent.uploadId) "upload intent replay was not idempotent"
  Send-Upload -Intent $intent -ContentType "application/pdf" -Bytes $pdf -Checksum $pdfChecksum
  Invoke-Api -Method Post -Path "/v1/resume-uploads/$($intent.uploadId)/complete" -Token $resumeIdentity.Token -Body $null | Out-Null
  $readyPdf = Wait-ResumeStatus -Identity $resumeIdentity -ResumeId $intent.resumeId -Expected @("ready")

  $docx = New-DocxFixture
  $docxChecksum = Get-Sha256 -Bytes $docx
  $docxIntent = New-UploadIntent -Identity $resumeIdentity -FileName "acceptance.docx" -ContentType "application/vnd.openxmlformats-officedocument.wordprocessingml.document" -Bytes $docx -Checksum $docxChecksum -IdempotencyKey "phase1-docx-$([Guid]::NewGuid())"
  Send-Upload -Intent $docxIntent -ContentType "application/vnd.openxmlformats-officedocument.wordprocessingml.document" -Bytes $docx -Checksum $docxChecksum
  Invoke-Api -Method Post -Path "/v1/resume-uploads/$($docxIntent.uploadId)/complete" -Token $resumeIdentity.Token -Body $null | Out-Null
  $readyDocx = Wait-ResumeStatus -Identity $resumeIdentity -ResumeId $docxIntent.resumeId -Expected @("ready")

  $invalid = [Text.Encoding]::ASCII.GetBytes("not-a-pdf-document")
  $invalidChecksum = Get-Sha256 -Bytes $invalid
  $invalidIntent = New-UploadIntent -Identity $resumeIdentity -FileName "invalid.pdf" -ContentType "application/pdf" -Bytes $invalid -Checksum $invalidChecksum -IdempotencyKey "phase1-invalid-$([Guid]::NewGuid())"
  Send-Upload -Intent $invalidIntent -ContentType "application/pdf" -Bytes $invalid -Checksum $invalidChecksum
  Invoke-Api -Method Post -Path "/v1/resume-uploads/$($invalidIntent.uploadId)/complete" -Token $resumeIdentity.Token -Body $null | Out-Null
  $invalidResult = Wait-ResumeStatus -Identity $resumeIdentity -ResumeId $invalidIntent.resumeId -Expected @("rejected")

  $expected = [Text.Encoding]::ASCII.GetBytes("%PDF-1.4`nexpected-A`n%%EOF`n")
  $actual = [Text.Encoding]::ASCII.GetBytes("%PDF-1.4`nexpected-B`n%%EOF`n")
  Assert-True ($expected.Length -eq $actual.Length) "checksum fixtures must have equal size"
  $expectedChecksum = Get-Sha256 -Bytes $expected
  $mismatchIntent = New-UploadIntent -Identity $resumeIdentity -FileName "mismatch.pdf" -ContentType "application/pdf" -Bytes $actual -Checksum $expectedChecksum -IdempotencyKey "phase1-mismatch-$([Guid]::NewGuid())"
  Send-Upload -Intent $mismatchIntent -ContentType "application/pdf" -Bytes $actual -Checksum $expectedChecksum
  Invoke-Api -Method Post -Path "/v1/resume-uploads/$($mismatchIntent.uploadId)/complete" -Token $resumeIdentity.Token -Body $null | Out-Null
  $mismatchResult = Wait-ResumeStatus -Identity $resumeIdentity -ResumeId $mismatchIntent.resumeId -Expected @("rejected")

  $eicarSignature = 'X5O!P%@AP[4\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*'
  $eicarDocument = "%PDF-1.7`n1 0 obj`n<< /Type /Catalog /Names << /EmbeddedFiles << /Names [(eicar.com) 2 0 R] >> >> >>`nendobj`n2 0 obj`n<< /Type /Filespec /F (eicar.com) /EF << /F 3 0 R >> >>`nendobj`n3 0 obj`n<< /Type /EmbeddedFile /Length 68 >>`nstream`n$eicarSignature`nendstream`nendobj`ntrailer << /Root 1 0 R >>`n%%EOF`n"
  $eicar = [Text.Encoding]::ASCII.GetBytes($eicarDocument)
  $eicarChecksum = Get-Sha256 -Bytes $eicar
  $eicarIntent = New-UploadIntent -Identity $resumeIdentity -FileName "eicar.pdf" -ContentType "application/pdf" -Bytes $eicar -Checksum $eicarChecksum -IdempotencyKey "phase1-eicar-$([Guid]::NewGuid())"
  Send-Upload -Intent $eicarIntent -ContentType "application/pdf" -Bytes $eicar -Checksum $eicarChecksum
  Invoke-Api -Method Post -Path "/v1/resume-uploads/$($eicarIntent.uploadId)/complete" -Token $resumeIdentity.Token -Body $null | Out-Null
  $eicarResult = Wait-ResumeStatus -Identity $resumeIdentity -ResumeId $eicarIntent.resumeId -Expected @("rejected")

  $expiredIntent = New-UploadIntent -Identity $resumeIdentity -FileName "expired.pdf" -ContentType "application/pdf" -Bytes $pdf -Checksum $pdfChecksum -IdempotencyKey "phase1-expired-$([Guid]::NewGuid())"
  Invoke-DockerSql -Sql "UPDATE resume_uploads SET expires_at=now()-interval '1 minute' WHERE id='$($expiredIntent.uploadId)';"
  $expiredResult = Wait-ResumeStatus -Identity $resumeIdentity -ResumeId $expiredIntent.resumeId -Expected @("failed")

  Invoke-Api -Method Patch -Path "/v1/onboarding" -Token $resumeIdentity.Token -Body @{
    currentStep = 4; status = "completed"; profileMethod = "resume"
  } | Out-Null
  $resumeOnboarding = Invoke-Api -Method Get -Path "/v1/onboarding" -Token $resumeIdentity.Token -Body $null
  Assert-True ($resumeOnboarding.status -eq "completed" -and $resumeOnboarding.profileMethod -eq "resume") "resume onboarding did not persist"
  Assert-True ($readyPdf.status -eq "ready" -and $readyDocx.status -eq "ready") "valid documents were not accepted"
  Assert-True ($invalidResult.rejectionReason -match "signature") "invalid PDF rejection reason was not persisted"
  Assert-True ($mismatchResult.rejectionReason -match "checksum") "checksum mismatch was not rejected"
  Assert-True ($eicarResult.rejectionReason -match "FOUND") "malware test file was not rejected by ClamAV"
  Assert-True ($expiredResult.rejectionReason -match "expired") "expired upload was not failed"

  $outboxId = [Guid]::NewGuid().ToString()
  $aggregateId = [Guid]::NewGuid().ToString()
  $drainDeadline = (Get-Date).AddSeconds(60)
  do {
    $pendingTenantEvents = Invoke-DockerSqlScalar -Sql "SELECT count(*) FROM outbox_events WHERE organization_id IN ('$($manualSession.organization.id)','$($resumeSession.organization.id)') AND published_at IS NULL;"
    if ($pendingTenantEvents -eq "0") { break }
    Start-Sleep -Seconds 2
  } while ((Get-Date) -lt $drainDeadline)
  Assert-True ($pendingTenantEvents -eq "0") "tenant outbox did not drain before retry acceptance"
  & docker compose --env-file $EnvironmentFile stop localstack | Out-Null
  Assert-True ($LASTEXITCODE -eq 0) "LocalStack could not be stopped for retry acceptance"
  $localstackStopped = $true
  Invoke-DockerSql -Sql "INSERT INTO outbox_events(id,organization_id,event_type,aggregate_type,aggregate_id,payload) VALUES('$outboxId','$($resumeSession.organization.id)','acceptance.outbox.v1','acceptance','$aggregateId',jsonb_build_object('kind','phase1-acceptance'));"
  $retryDeadline = (Get-Date).AddSeconds(45)
  do {
    Start-Sleep -Seconds 2
    $retryState = Invoke-DockerSqlScalar -Sql "SELECT attempts || ':' || (published_at IS NOT NULL)::text FROM outbox_events WHERE id='$outboxId';"
  } while ($retryState -eq "0:false" -and (Get-Date) -lt $retryDeadline)
  Assert-True ($retryState -match '^[1-9][0-9]*:false$') "outbox failure did not persist a retry attempt"
  & docker compose --env-file $EnvironmentFile up -d --wait localstack | Out-Null
  Assert-True ($LASTEXITCODE -eq 0) "LocalStack could not be restarted"
  Wait-LocalStackResources
  $localstackStopped = $false
  $publishDeadline = (Get-Date).AddSeconds(60)
  do {
    Start-Sleep -Seconds 2
    $published = Invoke-DockerSqlScalar -Sql "SELECT (published_at IS NOT NULL)::text FROM outbox_events WHERE id='$outboxId';"
  } while ($published -ne "true" -and (Get-Date) -lt $publishDeadline)
  Assert-True ($published -eq "true") "outbox event was not delivered after queue recovery"
  $queueDepth = & docker compose --env-file $EnvironmentFile exec -T localstack awslocal sqs get-queue-attributes --queue-url http://localhost:4566/000000000000/roliq-domain-events --attribute-names ApproximateNumberOfMessages --query 'Attributes.ApproximateNumberOfMessages' --output text
  Assert-True ($LASTEXITCODE -eq 0 -and [int](($queueDepth | Out-String).Trim()) -gt 0) "LocalStack SQS did not retain delivered domain events"

  [pscustomobject]@{
    ClerkOIDC              = "PASS"
    BootstrapIdempotency   = "PASS"
    TenantIsolation        = "PASS"
    ManualOnboarding       = "PASS"
    ResumeOnboarding       = "PASS"
    PersistedDashboard     = "PASS"
    ValidPdf               = "PASS"
    ValidDocx              = "PASS"
    InvalidSignature       = "PASS"
    ChecksumMismatch       = "PASS"
    MalwareDetection       = "PASS"
    UploadExpiration       = "PASS"
    UploadIntentIdempotency = "PASS"
    DatabaseRlsIsolation   = "PASS"
    OutboxRetryAndDelivery = "PASS"
  } | Format-List
}
finally {
  if ($localstackStopped) {
    try {
      & docker compose --env-file $EnvironmentFile up -d --wait localstack | Out-Null
      Wait-LocalStackResources
    }
    catch { Write-Warning $_ }
  }
  foreach ($organizationId in $organizations) {
    if ($organizationId) {
      try {
        Invoke-DockerSql -Sql "DELETE FROM resume_versions WHERE organization_id='$organizationId'; DELETE FROM organizations WHERE id='$organizationId';"
      }
      catch { Write-Warning $_ }
      try { & docker compose --env-file $EnvironmentFile exec -T localstack awslocal s3 rm "s3://roliq-resumes-local/quarantine/$organizationId" --recursive | Out-Null } catch { Write-Warning $_ }
    }
  }
  foreach ($identity in $identities) {
    if ($identity.UserId) {
      try { Invoke-DockerSql -Sql "DELETE FROM users WHERE id=(SELECT user_id FROM auth_identities WHERE subject='$($identity.UserId)');" } catch { Write-Warning $_ }
      try { Invoke-Clerk -Method Delete -Path "/users/$($identity.UserId)" -Headers $clerkHeaders -Body $null | Out-Null } catch { Write-Warning $_ }
    }
  }
}
