package kubenav

import io.circe.Json
import io.circe.generic.auto._
import io.circe.syntax._
import io.circe.yaml.syntax._
import io.k8s.apimachinery.pkg.apis.meta.v1.Time

import scala.reflect.ClassTag
import scala.reflect.runtime.universe.runtimeMirror

/**
 * Serialize arbitrary objects to avoid requiring a static list of k8s objects.
 */
object Serializer {
  val mirror = runtimeMirror(getClass.getClassLoader)

  def serialize[T: ClassTag](obj: T): List[String] = {
    val symbol = mirror.classSymbol(obj.getClass)
    if (!symbol.isCaseClass) return List(s"Reflection based serialization has only been tested with case classes: ${obj.getClass.getCanonicalName}")

    // This is reasonably likely to blow up. Try it and log errors.
    List(jsonTraverse(obj).deepDropNullValues.asYaml.spaces4)
  }

  def jsonTraverse[T: ClassTag](obj: T): Json = {
    obj match {
      // All of the maps are String,String so this should work
      case m: scala.collection.Map[String, String] => m.asJson
      case s: scala.collection.Seq[Any]            => s.map(jsonTraverse).asJson
      case b: Boolean                              => b.asJson
      case i: Int                                  => i.asJson
      case f: Float                                => f.asJson
      case d: Double                               => d.asJson
      case l: Long                                 => l.asJson
      case s: String                               => s.asJson
      case t: Time                                 => t.asJson
      case None                                    => Json.Null
      case Some(any)                               => jsonTraverse(any)
      case _ =>
        Json.fromFields(jsonObject(obj))
    }
  }

  def jsonObject[T: ClassTag](obj: T): Iterable[(String, Json)] = {
    val symbol = mirror.classSymbol(obj.getClass)
    lazy val instanceMirror = mirror.reflect(obj)
    symbol.info.decls.toList
      .filter(decl => decl.isPublic && decl.isMethod)
      .map(_.asMethod)
      .filter(_.isGetter)
      .map { method =>
        val v = instanceMirror.reflectMethod(method)()
        (method.name.toString, jsonTraverse(v))
      }
  }
}
