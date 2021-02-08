
## Develop
1. export JAVA_HOME="$(/usr/libexec/java_home -v "11")"
1. brew install sbt kind
1. make cluster-create run

## Useful Links
* [k8s object api](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/)
* [ZIO type aliases](https://zio.dev/docs/overview/overview_index#type-aliases)
* [ZIO modules](https://zio.dev/docs/howto/howto_use_layers#the-zlayer-data-type)

## Native
1. brew install --cask graalvm/tap/graalvm-ce-java11
1. xattr -r -d com.apple.quarantine /Library/Java/JavaVirtualMachines/graalvm-ce-java11-21.0.0
1. /Library/Java/JavaVirtualMachines/graalvm-ce-java11-21.0.0/Contents/Home/bin/gu install native-image