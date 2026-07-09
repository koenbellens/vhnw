# VHNW Android-miner (Laag 4a)

Een kleine Android-app die xmrig als **foreground service** draait, zodat minen
niet meer stopt zoals bij een gewone achtergrond-app (het probleem dat de
Termux/XMRig-app-aanpak destijds had). Werkt op elke telefoon vanaf Android 7.0
(minSdk 24), zonder root, en meldt zich — net als de Pi/laptop-agent — bij
hash.vbnw.nl.

## Wat nog ontbreekt: het xmrig-binary

Deze app start xmrig als los proces (`applicationInfo.nativeLibraryDir/libxmrig.so`),
precies zoals de Go-agent dat op de Pi/laptop doet. Er zit **geen xmrig-binary
bij** in deze repo — dat bouwen we zelf uit de officiële broncode via de Android
NDK, in plaats van een willekeurig precompiled binary van internet te
vertrouwen (een miner-executable die je zomaar downloadt is een voor de hand
liggend doelwit voor een gemanipuleerde/trojanized versie).

Stappen (met Android Studio + NDK, die installeert automatisch als je een
native project opent — of via `sdkmanager --install "ndk;27.0.12077973"`):

```bash
git clone --recursive https://github.com/xmrig/xmrig.git
cd xmrig
mkdir build && cd build
cmake .. \
  -DCMAKE_TOOLCHAIN_FILE=$ANDROID_NDK/build/cmake/android.toolchain.cmake \
  -DANDROID_ABI=arm64-v8a \
  -DANDROID_PLATFORM=android-24 \
  -DXMRIG_DEPS=<pad naar vooraf gebouwde deps: openssl, libuv, hwloc> \
  -DWITH_HWLOC=OFF   # simpelste eerste build: hwloc is optioneel, kost topologie-detectie
cmake --build . -j
```

Zet het resultaat (`xmrig`) daarna als:

```
android/app/src/main/jniLibs/arm64-v8a/libxmrig.so
```

(De naam moet met `lib` beginnen en op `.so` eindigen — dat is de officiële
Android-truc om een los uitvoerbaar bestand mee te installeren zonder root:
Android pakt alles in `jniLibs` uit naar de eigen, uitvoerbare map van de app.)

De meeste oude telefoons zijn **arm64-v8a** (64-bit ARM); voor hele oude
32-bit-toestellen bouw je hetzelfde met `-DANDROID_ABI=armeabi-v7a` en zet je
het resultaat in `jniLibs/armeabi-v7a/`.

## Bouwen en installeren

1. Open de map `android/` in Android Studio ("Open" → kies deze map, niet de
   hele repo).
2. Zie je bij het openen een melding over een ontbrekende Gradle-wrapper? Kies
   "OK"/"Use Gradle from: Android Studio default" — Android Studio genereert
   dan zelf de wrapper-bestanden.
3. Zet je telefoon in **Developer options → USB debugging** aan en verbind 'm
   met USB.
4. Run ▶ in Android Studio → kies je telefoon → de app installeert en opent.

## Batterij-instellingen die je ná installatie nog moet zetten

Stock Android vraagt de app zelf om vrijstelling (de knop "Batterijoptimalisatie
uitzetten" in de app doet dit). Maar veel merken hebben er zelf nóg een laag
bovenop:

- **Xiaomi/MIUI**: Instellingen → Apps → VHNW Miner → Batterij → "Geen
  restricties" + Beveiliging-app → Autostart aanzetten.
- **Huawei/Honor**: Instellingen → Batterij → App-lancering → VHNW Miner op
  "Handmatig beheren" met alle 3 schakelaars aan.
- **Samsung**: Instellingen → Apps → VHNW Miner → Batterij → "Niet
  geoptimaliseerd", en zet 'm niet in de "Slapende apps"-lijst.

Zonder deze stap lijkt de app te werken maar stopt hij na een paar uur — precies
het oorspronkelijke probleem, dus niet overslaan op een telefoon die je écht
24/7 wil laten meedraaien.

## Dashboard

`device_type` staat vast op `"cpu-agent"` (zelfde badge als de Pi/laptop). Wil
je telefoons een eigen badge geven op hash.vbnw.nl, dan is dat een kleine
uitbreiding van de hub later (`internal/hub/store.go` + de PHP-versie).
