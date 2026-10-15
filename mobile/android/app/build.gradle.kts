plugins {
    id("com.android.application")
    // The Flutter Gradle Plugin must be applied after the Android and Kotlin Gradle plugins.
    id("dev.flutter.flutter-gradle-plugin")
}

android {
    namespace = "mx.romodevx.control_contable"
    compileSdk = flutter.compileSdkVersion
    ndkVersion = flutter.ndkVersion

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }

    defaultConfig {
        applicationId = "mx.romodevx.control_contable"
        minSdk = flutter.minSdkVersion
        targetSdk = flutter.targetSdkVersion
        // Uses the version code from pubspec.yaml. When using split APKs, 1000 * ABI_VERSION
        // is added automatically by Flutter. (https://developer.android.com/studio/build/configure-apk-splits#configure-APK-versions)
        // You can force using the value of versionCode by specifying the `-P force-version-code-ignoring-abi=true`
        // flag during build.
        versionCode = flutter.versionCode
        versionName = flutter.versionName
    }

    // Per-flavor app names below use resValue, off by default in recent AGP.
    buildFeatures {
        resValues = true
    }

    // One product flavor per Dart entrypoint (lib/main_<flavor>.dart). dev and
    // staging install side by side with prod.
    flavorDimensions += "environment"
    productFlavors {
        create("dev") {
            dimension = "environment"
            applicationIdSuffix = ".dev"
            versionNameSuffix = "-dev"
            resValue("string", "app_name", "Control Contable Dev")
        }
        create("staging") {
            dimension = "environment"
            applicationIdSuffix = ".staging"
            versionNameSuffix = "-staging"
            resValue("string", "app_name", "Control Contable Staging")
        }
        create("prod") {
            dimension = "environment"
            resValue("string", "app_name", "Control Contable")
        }
    }

    buildTypes {
        release {
            // Store signing arrives with REL-04 (OB7); debug keys keep
            // `flutter run --release` working until then.
            signingConfig = signingConfigs.getByName("debug")
            // R8 shrink + obfuscation of the Java/Kotlin side. The Dart side is
            // obfuscated by `flutter build --obfuscate --split-debug-info`.
            isMinifyEnabled = true
            isShrinkResources = true
        }
    }
}

// Push (M9-09). The Firebase config is never committed (public repo, OB9): CI writes it from a
// secret, a developer drops their own in. Without one the plugin is not applied and every build
// works; the app then reports push as not available. One android/app/google-services.json can
// carry the three flavors' package names, or one per flavor in src/<flavor>/ (README).
val firebaseConfigs = fileTree("src") { include("*/google-services.json") } +
    files("google-services.json").filter { it.exists() }
if (!firebaseConfigs.isEmpty) {
    apply(plugin = "com.google.gms.google-services")
}

kotlin {
    compilerOptions {
        jvmTarget = org.jetbrains.kotlin.gradle.dsl.JvmTarget.JVM_17
    }
}

flutter {
    source = "../.."
}
