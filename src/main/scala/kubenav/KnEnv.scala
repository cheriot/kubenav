package kubenav
import zio.logging._

object KnEnv {
  val defaultLogLevel = LogLevel.Fatal
  val logLevels = List(
    LogLevel.Off,
    LogLevel.Trace,
    LogLevel.Debug,
    LogLevel.Info,
    LogLevel.Warn,
    LogLevel.Error,
    LogLevel.Fatal
  )

  def loggingLayer(logLevel: LogLevel) =
    Logging.consoleErr(
      logLevel = logLevel,
      format = LogFormat.ColoredLogFormat()
    ) >>> Logging.withRootLoggerName("kubenav-cli")
}
