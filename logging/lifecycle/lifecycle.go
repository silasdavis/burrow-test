// Copyright 2017 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lifecycle

// No package in ./logging/... should depend on lifecycle
import (
	"os"

	"time"

	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/adapters/stdlib"
	tmLog15adapter "github.com/hyperledger/burrow/logging/adapters/tendermint_log15"
	"github.com/hyperledger/burrow/logging/config"
	"github.com/hyperledger/burrow/logging/loggers"
	"github.com/hyperledger/burrow/logging/structure"

	kitlog "github.com/go-kit/kit/log"
	"github.com/hyperledger/burrow/logging/types"
	"github.com/streadway/simpleuuid"
	tmLog15 "github.com/tendermint/log15"
)

// Lifecycle provides a canonical source for burrow loggers. Components should use the functions here
// to set up their root logger and capture any other logging output.

// Obtain a logger from a LoggingConfig
func NewLoggerFromLoggingConfig(loggingConfig *config.LoggingConfig) (types.InfoTraceLogger, error) {
	if loggingConfig == nil {
		return NewStdErrLogger(), nil
	}
	outputLogger, err := infoTraceLoggerFromLoggingConfig(loggingConfig)
	if err != nil {
		return nil, err
	}
	return NewLogger(outputLogger), nil
}

// Hot swap logging config by replacing output loggers of passed InfoTraceLogger
// with those built from loggingConfig
func SwapOutputLoggersFromLoggingConfig(logger types.InfoTraceLogger,
	loggingConfig *config.LoggingConfig) error {
	outputLogger, err := infoTraceLoggerFromLoggingConfig(loggingConfig)
	if err != nil {
		return err
	}
	logger.SwapOutput(outputLogger)
	return nil
}

func NewStdErrLogger() types.InfoTraceLogger {
	logger := loggers.NewStreamLogger(os.Stderr, "terminal")
	return NewLogger(logger)
}

// Provided a standard logger that outputs to the supplied underlying outputLogger
func NewLogger(outputLogger kitlog.Logger) types.InfoTraceLogger {
	infoTraceLogger := loggers.NewInfoTraceLogger(outputLogger)
	// Create a random ID based on start time
	uuid, _ := simpleuuid.NewTime(time.Now())
	var runId string
	if uuid != nil {
		runId = uuid.String()
	}
	return logging.WithMetadata(infoTraceLogger.With(structure.RunId, runId))
}

func CaptureTendermintLog15Output(infoTraceLogger types.InfoTraceLogger) {
	tmLog15.Root().SetHandler(
		tmLog15adapter.InfoTraceLoggerAsLog15Handler(infoTraceLogger.
			With(structure.CapturedLoggingSourceKey, "tendermint_log15")))
}

func CaptureStdlibLogOutput(infoTraceLogger types.InfoTraceLogger) {
	stdlib.CaptureRootLogger(infoTraceLogger.
		With(structure.CapturedLoggingSourceKey, "stdlib_log"))
}

// Helpers
func infoTraceLoggerFromLoggingConfig(loggingConfig *config.LoggingConfig) (kitlog.Logger, error) {
	outputLogger, _, err := loggingConfig.RootSink.BuildLogger()
	if err != nil {
		return nil, err
	}
	return outputLogger, nil
}
