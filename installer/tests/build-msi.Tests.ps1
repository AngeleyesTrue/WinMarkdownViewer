# Pester tests for build-msi.ps1
# Validates script structure, parameters, and prerequisites
# Compatible with Pester 3.4+ and 5.x

$BuildScriptPath = Join-Path (Join-Path $PSScriptRoot "..") "build-msi.ps1"
$WixDir = Join-Path (Join-Path $PSScriptRoot "..") "wix"

Describe "build-msi.ps1 Script Validation" {

    Context "Script file" {
        It "should exist" {
            $BuildScriptPath | Should Exist
        }

        It "should be valid PowerShell" {
            $errors = $null
            [System.Management.Automation.PSParser]::Tokenize(
                (Get-Content $BuildScriptPath -Raw), [ref]$errors
            )
            $errors.Count | Should Be 0
        }
    }

    Context "Script parameters" {
        $ScriptInfo = Get-Command $BuildScriptPath
        $Parameters = $ScriptInfo.Parameters

        It "should have Version parameter" {
            $Parameters.ContainsKey('Version') | Should Be $true
        }

        It "should have Configuration parameter" {
            $Parameters.ContainsKey('Configuration') | Should Be $true
        }

        It "should have OutputDir parameter" {
            $Parameters.ContainsKey('OutputDir') | Should Be $true
        }

        It "should have SkipGoBuild parameter" {
            $Parameters.ContainsKey('SkipGoBuild') | Should Be $true
        }

        It "Version parameter should be string type" {
            $Parameters['Version'].ParameterType.Name | Should Be 'String'
        }

        It "Configuration parameter should be string type" {
            $Parameters['Configuration'].ParameterType.Name | Should Be 'String'
        }

        It "Configuration parameter should validate Release and Debug" {
            $validateSet = $Parameters['Configuration'].Attributes |
                Where-Object { $_ -is [System.Management.Automation.ValidateSetAttribute] }
            $validateSet | Should Not BeNullOrEmpty
            $validateSet.ValidValues -contains 'Release' | Should Be $true
            $validateSet.ValidValues -contains 'Debug' | Should Be $true
        }

        It "OutputDir parameter should be string type" {
            $Parameters['OutputDir'].ParameterType.Name | Should Be 'String'
        }

        It "SkipGoBuild parameter should be switch type" {
            $Parameters['SkipGoBuild'].ParameterType.Name | Should Be 'SwitchParameter'
        }
    }

    Context "Default values" {
        $Content = Get-Content $BuildScriptPath -Raw

        It "Configuration should default to Release" {
            $Content | Should Match '\$Configuration\s*=\s*"Release"'
        }

        It "OutputDir should default to ./dist" {
            $Content | Should Match '\$OutputDir\s*=\s*"\./dist"'
        }
    }

    Context "WiX source files" {
        It "WiX directory should exist" {
            $WixDir | Should Exist
        }

        It "should contain Components.wxs" {
            Join-Path $WixDir "Components.wxs" | Should Exist
        }

        It "should contain Directories.wxs" {
            Join-Path $WixDir "Directories.wxs" | Should Exist
        }

        It "should contain Registry.wxs" {
            Join-Path $WixDir "Registry.wxs" | Should Exist
        }

        It "should contain Shortcuts.wxs" {
            Join-Path $WixDir "Shortcuts.wxs" | Should Exist
        }

        It "should contain Package.wxs" {
            Join-Path $WixDir "Package.wxs" | Should Exist
        }

        It "should contain Variables.wxi" {
            Join-Path $WixDir "Variables.wxi" | Should Exist
        }
    }

    Context "Script content requirements" {
        $Content = Get-Content $BuildScriptPath -Raw

        It "should use strict mode" {
            $Content | Should Match 'Set-StrictMode'
        }

        It "should set ErrorActionPreference to Stop" {
            $Content | Should Match "ErrorActionPreference\s*=\s*'Stop'"
        }

        It "should reference go build command" {
            $Content | Should Match 'go build'
        }

        It "should reference wix build command" {
            $Content | Should Match 'wix.*build'
        }

        It "should use -ldflags with windowsgui" {
            $Content | Should Match 'windowsgui'
        }

        It "should verify MSI output exists" {
            $Content | Should Match 'Test-Path.*msi'
        }
    }
}
