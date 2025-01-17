package metaparser

import (
	"errors"
	"fmt"

	"github.com/dashwave/go-xcode/v2/artifacts"
	"github.com/dashwave/go-xcode/v2/zip"
)

// MacOSProjectIsNotSupported ...
var MacOSProjectIsNotSupported = errors.New("macOS project is not supported")

// ParseXCArchiveData ...
func (m *Parser) ParseXCArchiveData(pth string) (*ArtifactMetadata, error) {

	appInfo, scheme, err := m.readXCArchiveDeploymentMeta(pth)
	if err != nil {
		return &ArtifactMetadata{
			AppInfo: appInfo,
			Scheme:  scheme,
		}, fmt.Errorf("failed to parse deployment info for %s: %w", pth, err)
	}

	fileSize, err := m.fileManager.FileSizeInBytes(pth)
	if err != nil {
		m.logger.Warnf("Failed to get apk size, error: %s", err)
	}

	return &ArtifactMetadata{
		AppInfo:       appInfo,
		FileSizeBytes: fileSize,
		Scheme:        scheme,
	}, nil
}

func (m *Parser) readXCArchiveDeploymentMeta(pth string) (Info, string, error) {
	reader, err := zip.NewDefaultReader(pth, m.logger)
	if err != nil {
		return Info{}, "", err
	}
	defer func() {
		if err := reader.Close(); err != nil {
			m.logger.Warnf("%s", err)
		}
	}()

	xcarchiveReader := artifacts.NewXCArchiveReader(reader)
	isMacos := xcarchiveReader.IsMacOS()
	if isMacos {
		return Info{}, "", MacOSProjectIsNotSupported
	}
	archiveInfoPlist, err := xcarchiveReader.InfoPlist()
	if err != nil {
		return Info{}, "", fmt.Errorf("failed to unwrap Info.plist from xcarchive: %w", err)
	}

	iosXCArchiveReader := artifacts.NewIOSXCArchiveReader(reader)
	appInfoPlist, err := iosXCArchiveReader.AppInfoPlist()
	if err != nil {
		return Info{}, "", fmt.Errorf("failed to unwrap application Info.plist from xcarchive: %w", err)
	}

	appTitle, _ := appInfoPlist.GetString("CFBundleName")
	bundleID, _ := appInfoPlist.GetString("CFBundleIdentifier")
	version, _ := appInfoPlist.GetString("CFBundleShortVersionString")
	buildNumber, _ := appInfoPlist.GetString("CFBundleVersion")
	minOSVersion, _ := appInfoPlist.GetString("MinimumOSVersion")
	deviceFamilyList, _ := appInfoPlist.GetUInt64Array("UIDeviceFamily")
	scheme, _ := archiveInfoPlist.GetString("SchemeName")

	appInfo := Info{
		AppTitle:         appTitle,
		BundleID:         bundleID,
		Version:          version,
		BuildNumber:      buildNumber,
		MinOSVersion:     minOSVersion,
		DeviceFamilyList: deviceFamilyList,
	}

	return appInfo, scheme, nil
}
