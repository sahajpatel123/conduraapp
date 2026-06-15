; Synaptic GUI — minimal NSIS installer for Windows releases.
!include "MUI2.nsh"

Name "Synaptic"
OutFile "${OUTFILE}"
InstallDir "$PROGRAMFILES64\Synaptic"
RequestExecutionLevel admin
Unicode true

!define MUI_ABORTWARNING
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_LANGUAGE "English"

Section "Synaptic" SecMain
  SetOutPath "$INSTDIR"
  File "${EXE}"
  CreateShortcut "$DESKTOP\Synaptic.lnk" "$INSTDIR\synaptic.exe"
  WriteUninstaller "$INSTDIR\Uninstall.exe"
SectionEnd

Section "Uninstall"
  Delete "$INSTDIR\synaptic.exe"
  Delete "$INSTDIR\Uninstall.exe"
  Delete "$DESKTOP\Synaptic.lnk"
  RMDir "$INSTDIR"
SectionEnd
