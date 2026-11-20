package repository

import (
	"context"
	"database/sql/driver"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestUsageLogRepositoryAverageOutputTokenSpeed(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &usageLogRepository{sql: db}
	start := time.Date(2026, 11, 19, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 1)
	filters := usagestats.UsageLogFilters{GroupID: 7, StartTime: &start, EndTime: &end}
	query := `SELECT 1000.0 * SUM(output_tokens) / NULLIF(SUM(duration_ms - first_token_ms), 0) FROM usage_logs WHERE group_id = $1 AND created_at >= $2 AND created_at < $3 AND (request_type = $4 OR (request_type = 0 AND stream = TRUE AND openai_ws_mode = FALSE)) AND image_count = 0 AND output_tokens > 0 AND first_token_ms > 0 AND duration_ms > first_token_ms`
	for _, value := range []driver.Value{47.6, nil} {
		mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(int64(7), start, end, int16(service.RequestTypeStream)).
			WillReturnRows(sqlmock.NewRows([]string{"speed"}).AddRow(value))
		result, err := repo.GetAverageOutputTokenSpeed(context.Background(), filters)
		require.NoError(t, err)
		if value == nil {
			require.Nil(t, result)
		} else {
			require.NotNil(t, result)
			require.Equal(t, value, *result)
		}
	}
	require.NoError(t, mock.ExpectationsWereMet())
}
