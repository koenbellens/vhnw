plugins {
    id("com.android.application")
    id("org.jetbrains.kotlin.android")
}

android {
    namespace = "nl.vbnw.vhnw.miner"
    compileSdk = 34

    defaultConfig {
        applicationId = "nl.vbnw.vhnw.miner"
        minSdk = 24
        targetSdk = 34
        versionCode = 1
        versionName = "0.1.0"
    }

    buildTypes {
        release {
            isMinifyEnabled = false
        }
    }

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }
    kotlinOptions {
        jvmTarget = "17"
    }

    // xmrig wordt als natief "lib*.so"-bestand meegeleverd (zie android/README.md) zodat
    // Android het bij installatie uitpakt naar een uitvoerbare map van de app zelf —
    // geen root nodig om het te draaien.
    packaging {
        jniLibs.useLegacyPackaging = true
    }
}

dependencies {
    implementation("androidx.core:core-ktx:1.13.1")
    implementation("androidx.appcompat:appcompat:1.7.0")
    implementation("com.google.android.material:material:1.12.0")
}
