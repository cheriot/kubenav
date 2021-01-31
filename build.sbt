import Dependencies._

ThisBuild / scalaVersion     := "2.13.4"
ThisBuild / version          := "0.1.0-SNAPSHOT"
ThisBuild / organization     := "com.cheriot"
ThisBuild / organizationName := "cheriot"

val zioLoggingVersion = "0.5.6"

lazy val root = (project in file("."))
  .settings(
    name := "kubenav",
    libraryDependencies += scalaTest % Test,
    libraryDependencies += "dev.zio" %% "zio" % "1.0.4",
    libraryDependencies += "dev.zio" %% "zio-logging" % zioLoggingVersion
  )

// See https://www.scala-sbt.org/1.x/docs/Using-Sonatype.html for instructions on how to publish to Sonatype.
