package backup

import (
	"context"
	"time"

	"github.com/KaiserWerk/Tiny-Build-Server/internal/configuration"
	"github.com/KaiserWerk/sqldump"
)

func StartMakingBackups(ctx context.Context, appConfig *configuration.AppConfig) {
	uploader := sqldump.NewUploader(
		appConfig.StorageBox.Username,
		appConfig.StorageBox.Password,
		appConfig.StorageBox.Host,
	)

	uploader.ScheduleUpload(
		ctx,
		sqldump.GetMySQLBackupFileByDSN(appConfig.Database.DSN, "tiny-build-server"),
		true,
		".",
		30*time.Minute,
		4,
	)
}
