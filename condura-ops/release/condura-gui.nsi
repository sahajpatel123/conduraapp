; Condura GUI — minimal NSIS installer for Windows releases.
!include "MUI2.nsh"

Name "Condura"
OutFile "${OUTFILE}"
InstallDir "$PROGRAMFILES64\Condura"
RequestExecutionLevel admin
Unicode true

!define MUI_ABORTWARNING
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_LANGUAGE "English"

Section "Condura" SecMain
  SetOutPath "$INSTDIR"
  File "${EXE}"
  CreateShortcut "$DESKTOP\Condura.lnk" "$INSTDIR\condura.exe"
  WriteUninstaller "$INSTDIR\Uninstall.exe"
SectionEnd

Section "Uninstall"
  Delete "$INSTDIR\condura.exe"
  Delete "$INSTDIR\Uninstall.exe"
  Delete "$DESKTOP\Condura.lnk"
  RMDir "$INSTDIR"
SectionEnd
