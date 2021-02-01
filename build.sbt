import Dependencies._

ThisBuild / scalaVersion     := "2.13.4"
ThisBuild / version          := "0.1.0-SNAPSHOT"
ThisBuild / organization     := "com.cheriot"
ThisBuild / organizationName := "cheriot"

val zioLoggingVersion = "0.5.6"
val kubeClientVersion = "0.4.0"
val zioVersion = "1.0.4"
val zioCatsVersion = "2.2.0.1"

lazy val root = (project in file("."))
  .settings(
    name := "kubenav",
    libraryDependencies += scalaTest % Test,
    libraryDependencies += "dev.zio" %% "zio" % zioVersion,
    libraryDependencies += "dev.zio" %% "zio-test" % zioVersion,
    libraryDependencies += "dev.zio" %% "zio-test-sbt" % zioVersion,
    libraryDependencies += "dev.zio" %% "zio-interop-cats" % zioCatsVersion,
    libraryDependencies += "dev.zio" %% "zio-logging" % zioLoggingVersion,
    libraryDependencies += "com.goyeau" % "kubernetes-client_2.13" % kubeClientVersion
  )

// See https://www.scala-sbt.org/1.x/docs/Using-Sonatype.html for instructions on how to publish to Sonatype.