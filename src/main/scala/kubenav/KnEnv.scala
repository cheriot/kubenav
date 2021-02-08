package kubenav
import zio._
import zio.logging._

import cli.CommandLineParams
import kube.KubeClient

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

  def env(cliArgs: CommandLineParams): ZLayer[ZEnv, Nothing, KubeClient with Logging] = {

    val loggingLayer = KnEnv.loggingLayer(cliArgs.logLevel)
    val kubeClientLayer = loggingLayer >>> KubeClient.live
    loggingLayer ++ kubeClientLayer
  }
}
