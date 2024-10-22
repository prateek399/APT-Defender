package dao

import (
	"anti-apt-backend/config"
	"anti-apt-backend/extras"
	"fmt"
)

var FileHashesTable = "file_hashes"

func IsMalwareHashFromDb(md5 string, sha1 string, sha256 string) (bool, error) {
	queryString := fmt.Sprintf("SELECT count(*) FROM %s WHERE (md5 = '%s' or sha1 = '%s' or sha256 = '%s')", FileHashesTable, md5, sha1, sha256)

	count := int64(0)

	filehash := FileHashesRepo{
		QueryExecSet: []string{queryString},
		Result:       &count,
	}

	// slog.Println("MD5: ", md5)

	err := GormOperations(&filehash, config.Db, EXEC)
	if err != nil {
		return false, err
	}

	// slog.Println("HASH COUNT: ", count)

	if count > 0 {
		return true, nil
	}

	queryString = fmt.Sprintf("SELECT count(*) FROM %s WHERE (md5 = '%s' OR sha = '%s' OR sha256 = '%s') AND (final_verdict != '' OR final_verdict IS NOT NULL) AND final_verdict = 'block'", extras.FileOnDemandTable, md5, sha1, sha256)
	// slog.Println("QUERY STRING: ", queryString)
	count = 0
	fileOnDemand := DatabaseOperationsRepo{
		QueryExecSet: []string{queryString},
		Result:       &count,
	}

	err = GormOperations(&fileOnDemand, config.Db, EXEC)
	if err != nil {
		return false, err
	}

	// slog.Println("FOD COUNT: ", count)

	if count == 0 {
		return false, nil
	}

	// slog.Println("FINAL CALL: ", count)

	return true, nil
}

func IsCleanHashFromDb(md5 string, sha1 string, sha256 string) (bool, error) {
	queryString := fmt.Sprintf("SELECT count(*) FROM %s WHERE (md5 = '%s' OR sha = '%s' OR sha256 = '%s') AND (final_verdict != '' OR final_verdict IS NOT NULL)  AND final_verdict = 'allow'", extras.FileOnDemandTable, md5, sha1, sha256)

	count := int64(0)

	fileOnDemand := DatabaseOperationsRepo{
		QueryExecSet: []string{queryString},
		Result:       &count,
	}

	err := GormOperations(&fileOnDemand, config.Db, EXEC)
	if err != nil {
		return false, err
	}

	if count == 0 {
		return false, nil
	}

	return true, nil
}
