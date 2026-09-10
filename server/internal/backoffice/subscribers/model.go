package subscribers

import "github.com/leamout/leamout/internal/database/sqlc"

type Subscriber = sqlc.ListBackofficeSubscribersRow
type Detail = sqlc.GetBackofficeSubscriberRow
