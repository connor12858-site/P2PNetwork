# Setup the enviroment
param(
	[bool]$r= $false
)

Set-Location $PSScriptRoot
& ".\.venv\Scripts\Activate.ps1"

# Clear the dist folder
Remove-Item -Recurse -Force dist/*

# Build the new executable
pyinstaller app.py

# Copy the data folder over
New-Item -ItemType Directory -Force -Path dist/app/data
Copy-Item -Recurse -Force data/* dist/app/data/

# Check if we are running based on params
if ($r) {
	# Run the executable
	Start-Process -FilePath "$PSScriptRoot\dist\app\app.exe"
}

