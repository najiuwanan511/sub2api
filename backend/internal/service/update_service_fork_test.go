//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUpdateServiceRequiresForkAssetsAndChecksums(t *testing.T) {
	svc := NewUpdateService(nil, nil, "0.2.23", "release")
	assetName := "sub2api_0.2.24_" + svc.getArchiveName() + ".tar.gz"
	asset := Asset{Name: assetName, DownloadURL: "https://github.com/najiuwanan511/sub2api/releases/download/v0.2.24/" + assetName}
	require.NoError(t, validateDownloadURL(asset.DownloadURL))
	require.Error(t, validateDownloadURL("https://github.com/Wei-Shaw/sub2api/releases/download/v0.2.24/"+assetName))
	require.ErrorContains(t, svc.applyReleaseAssets(context.Background(), []Asset{asset}), "missing checksums.txt")
}
