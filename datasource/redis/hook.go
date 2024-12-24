package redis

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/xbmlz/webber/datasource"
)

type redisHook struct {
	config *Config
	logger datasource.Logger
}

type QueryLog struct {
	Query    string      `json:"query"`
	Duration int64       `json:"duration"`
	Args     interface{} `json:"args,omitempty"`
}

func (ql *QueryLog) PrettyPrint(writer io.Writer) {
	if ql.Query == "pipeline" {
		fmt.Fprintf(writer, "\u001B[38;5;8m%-32s \u001B[38;5;24m%-6s\u001B[0m %8d\u001B[38;5;8mµs\u001B[0m %s\n",
			clean(ql.Query), "REDIS", ql.Duration,
			ql.String()[1:len(ql.String())-1])
	} else {
		fmt.Fprintf(writer, "\u001B[38;5;8m%-32s \u001B[38;5;24m%-6s\u001B[0m %8d\u001B[38;5;8mµs\u001B[0m %v\n",
			clean(ql.Query), "REDIS", ql.Duration, ql.String())
	}
}

func clean(query string) string {
	query = regexp.MustCompile(`\s+`).ReplaceAllString(query, " ")
	query = strings.TrimSpace(query)

	return query
}

func (ql *QueryLog) String() string {
	if ql.Args == nil {
		return ""
	}

	switch args := ql.Args.(type) {
	case []interface{}:
		strArgs := make([]string, len(args))
		for i, arg := range args {
			strArgs[i] = fmt.Sprint(arg)
		}

		return strings.Join(strArgs, " ")
	default:
		return fmt.Sprint(ql.Args)
	}
}

func (r *redisHook) sendOperationStats(start time.Time, query string, args ...interface{}) {
	duration := time.Since(start).Microseconds()

	r.logger.Debug(&QueryLog{
		Query:    query,
		Duration: duration,
		Args:     args,
	})

}

// DialHook implements the redis.DialHook interface.
func (*redisHook) DialHook(next redis.DialHook) redis.DialHook {
	return next
}

// ProcessHook implements the redis.ProcessHook interface.
func (r *redisHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		start := time.Now()
		err := next(ctx, cmd)
		r.sendOperationStats(start, cmd.Name(), cmd.Args()...)

		return err
	}
}

// ProcessPipelineHook implements the redis.ProcessPipelineHook interface.
func (r *redisHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		start := time.Now()
		err := next(ctx, cmds)
		r.sendOperationStats(start, "pipeline", cmds[:len(cmds)-1])

		return err
	}
}
