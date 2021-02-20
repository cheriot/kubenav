import Dependencies._

ThisBuild / scalaVersion := "2.13.4"
ThisBuild / version := "0.1.0-SNAPSHOT"
ThisBuild / organization := "com.cheriot"
ThisBuild / organizationName := "cheriot"
ThisBuild / scalafixDependencies += "com.github.liancheng" %% "organize-imports" % "0.5.0"
ThisBuild / semanticdbEnabled := true
ThisBuild / semanticdbVersion := scalafixSemanticdb.revision
ThisBuild / scalacOptions ++= Seq(
  "-Wunused:imports", // required by `RemoveUnused` rule
  "-deprecation",
  "-feature",
  "-unchecked",
  "-language:postfixOps",
  "-language:higherKinds",
)

val zioLoggingVersion = "0.5.6"
val kubeClientVersion = "0.4.0-1-81fa69b"
val zioVersion = "1.0.4"
val zioCatsVersion = "2.2.0.1"
val circeVersion = "0.13.0"

lazy val root = (project in file("."))
  .enablePlugins(BuildInfoPlugin)
  .enablePlugins(ScalaJSPlugin)
  .settings(
    buildInfoKeys := Seq[BuildInfoKey](name, version, scalaVersion, sbtVersion),
    buildInfoPackage := "kubenav",
  )
  .settings(
    name := "kubenav",

    scalaJSUseMainModuleInitializer := true,
    libraryDependencies += "org.scala-js" %% "scalajs-stubs" % "1.0.0" % "provided",

    libraryDependencies += "dev.zio" %%% "zio" % zioVersion,
    libraryDependencies += "dev.zio" %%% "zio-interop-cats" % zioCatsVersion,
    libraryDependencies += "dev.zio" %%% "zio-logging" % zioLoggingVersion,
    libraryDependencies += "dev.zio" %% "zio-logging-slf4j" % zioLoggingVersion,
    libraryDependencies += "dev.zio" %% "zio-logging-slf4j-bridge" % zioLoggingVersion,

    libraryDependencies += ("com.goyeau" %%% "kubernetes-client" % kubeClientVersion).excludeAll(
      // Remove the slf4j backend so zio-logging-slf4j-bridge can feed them into zio-logging.
      ExclusionRule(organization = "ch.qos.logback")
    ),
    libraryDependencies += "org.typelevel" %%% "cats-effect" % "2.2.0" withSources () withJavadoc (),
    libraryDependencies ++= Seq(
      "io.circe" %%% "circe-core",
      "io.circe" %%% "circe-generic",
      "io.circe" %%% "circe-parser",
    ).map(_ % circeVersion),
    libraryDependencies += "io.circe" %% "circe-yaml" % "0.13.1", // This is published for circeVersion. Why is sbt not finding it?

    libraryDependencies += "com.github.scopt" %%% "scopt" % "4.0.0",

    libraryDependencies += "dev.zio" %%% "zio-test" % zioVersion % "test",
    libraryDependencies += "dev.zio" %%% "zio-test-sbt" % zioVersion % "test",

    testFrameworks += new TestFramework("zio.test.sbt.ZTestFramework"),
  )

// See https://www.scala-sbt.org/1.x/docs/Using-Sonatype.html for instructions on how to publish to Sonatype.
