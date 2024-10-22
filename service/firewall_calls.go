package service

import (
	"anti-apt-backend/dao"
	"anti-apt-backend/extras"
	"strings"
)

func FileFromFireWall(hashed string) int {
	hashed = strings.TrimSpace(hashed)
	// slog.Println("FILE HASH: ", hashed)
	if len(hashed) > 0 {
		isMalware, err := dao.IsMalwareHashFromDb(hashed, hashed, hashed)
		if err != nil {
			// slog.Println("Error: ", err)
			return extras.FW_EMPTY
		}

		isClean, err := dao.IsCleanHashFromDb(hashed, hashed, hashed)
		if err != nil {
			// slog.Println("Error: ", err)
			return extras.FW_EMPTY
		}

		if isClean {
			// slog.Println("CLEAN VERDICT")
			return extras.FW_CLEAN
		}

		if isMalware {
			// slog.Println("BLOCK VERDICT")
			return extras.FW_BLOCK
		}
	}

	// slog.Println("EMPTY")
	return extras.FW_EMPTY
}

// func FetchJobIDForFw(jobID string) int {
// 		var fod map[string]model.FileOnDemand
// 		var uod map[string]model.UrlOnDemand
// 		var err error
// 		fod, err = dao.FetchFileOnDemandProfile(map[string]any{"JobID": jobID})
// 		if err != nil && err != extras.ErrNoRecordForFileOnDemand {
// 			return extras.FW_RETRY
// 		}
// 		if err == extras.ErrNoRecordForFileOnDemand {
// 			uod, err = dao.FetchUrlOnDemandProfile(map[string]any{"JobID": jobID})
// 			if err != nil {
// 				return extras.FW_RETRY
// 			}
// 			val := uod[jobID]
// 			if val.Status != extras.REPORTED {
// 				return extras.FW_UNKNOWN
// 			}
// 			if val.FinalVerdict == extras.ALLOW {
// 				return extras.FW_CLEAN
// 			}
// 			return extras.FW_BLOCK
// 		}

// 	val := fod[jobID]

// 		if val.Status != extras.REPORTED {
// 			return extras.FW_UNKNOWN
// 		}

// 		if val.FinalVerdict == extras.ALLOW {
// 			return extras.FW_CLEAN
// 		} else {

// 			return extras.FW_BLOCK
// 		}

// 	return extras.FW_EMPTY // Remove this condition just used for now
// }
