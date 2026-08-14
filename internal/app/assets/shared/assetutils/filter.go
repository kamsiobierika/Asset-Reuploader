package assetutils

import (
	"fmt"

	"github.com/kartFr/Asset-Reuploader/internal/app/context"
	"github.com/kartFr/Asset-Reuploader/internal/app/request"
	"github.com/kartFr/Asset-Reuploader/internal/roblox/develop"
)

func NewFilter(ctx *context.Context, r *request.Request, assetTypeID int32) func(assetsInfo develop.GetAssetsInfoResponse) []*develop.AssetInfo {
	creatorID := r.CreatorID
	userID := ctx.Client.UserInfo.ID
	checkUserID := !r.IsGroup

	return func(assetsInfo develop.GetAssetsInfoResponse) []*develop.AssetInfo {
		filteredAssetsInfo := assetsInfo.Data[:0]
		typeMismatchCount := 0
		for _, info := range assetsInfo.Data {
			if info.TypeID != assetTypeID {
				typeMismatchCount++
				if typeMismatchCount <= 3 && ctx.DebugWriter != nil {
					debugMsg := fmt.Sprintf("Asset %d (type: %s, typeID: %d) doesn't match expected typeID %d", info.ID, info.Type, info.TypeID, assetTypeID)
					ctx.DebugWriter.WriteLine(debugMsg)
				}
				continue
			}

			assetCreatorID := info.Creator.TargetID
			if assetCreatorID == creatorID || assetCreatorID == 1 || (checkUserID && assetCreatorID == userID) {
				continue
			}

			filteredAssetsInfo = append(filteredAssetsInfo, info)
		}
		if typeMismatchCount > 3 && ctx.DebugWriter != nil {
			ctx.DebugWriter.WriteLine(fmt.Sprintf("... and %d more assets with mismatched typeID", typeMismatchCount-3))
		}
		return filteredAssetsInfo
	}
}
