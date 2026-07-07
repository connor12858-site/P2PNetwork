Set-Location $PSScriptRoot

& ".\.venv\Scripts\Activate.ps1"

$configPath = Join-Path $PSScriptRoot "data\config.yaml"
$host = "0.0.0.0"
$port = 8000

foreach ($line in Get-Content $configPath) {
	$trimmed = $line.Trim()

	if ($trimmed -match '^host:\s*(.+)$') {
		$host = $Matches[1].Trim()
	}

	if ($trimmed -match '^port:\s*(\d+)$') {
		$port = [int]$Matches[1]
	}
}

fastapi run main.py --host $host --port $port