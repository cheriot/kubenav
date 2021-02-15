package kubenav

import zio._
import zio.console._
import zio.test.Assertion._
import zio.test._
import zio.test.environment._

import HelloWorld._

object HelloWorld {
  def sayHello: URIO[Console, Unit] =
    console.putStrLn("Hello, World!")
}

object HelloWorldSpec extends DefaultRunnableSpec {
  val namespace = "hello"
  val resourceType = "service"
  val resourceName = "hello-svc"
  val relation = "deployment"

  def spec = suite("All suites")(helloSuite, entrypointSuite)

  val helloSuite = suite("HelloWorldSpec")(
    testM("sayHello correctly displays output") {
      for {
        _ <- sayHello
        output <- TestConsole.output
      } yield assert(output)(equalTo(Vector("Hello, World!\n")))
    }
  )

  val entrypointSuite = suite("EntrypointSpec")(
    test("find implicit instance dynamically") {
      // val deployment = get("hello", "deployment", "hellp-dep")
      assert(false)(isFalse)
    }
  )
}
