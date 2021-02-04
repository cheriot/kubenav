package kubenav

import org.rogach.scallop._
import zio.logging._

import KnEnv._

package object cli {

  case class CommandLineArgs(
    logLevel: LogLevel
  )

  def parse(args: List[String]) = {
    val parsedArgs = new cli.CommandLineConf(args)
    parsedArgs.verify()

    val logLevel = parsedArgs.logLevel.toOption
      .flatMap(matchLogLevel)
      // Ignore this error. Scallop will validate and default the user provided levelName
      .getOrElse(defaultLogLevel)

    CommandLineArgs(logLevel = logLevel)
  }

  class CommandLineConf(args: Seq[String]) extends ScallopConf(args) {
    val logLevel = choice(
      default = Some(defaultLogLevel.render),
      choices = logLevels.map(_.render),
      descr = "Log level of kubenav client."
    )
  }

  def matchLogLevel(levelName: String): Option[LogLevel] =
    logLevels.find(_.render == levelName)
}
