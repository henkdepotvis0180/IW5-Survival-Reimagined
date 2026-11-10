# Survival Reimagined Launcher

Eigen launcher voor Windows 10/11 (64-bit), met de huidige modbestanden ingebouwd.
Geen herstel van de gesloten launcher van de oorspronkelijke ontwikkelaar.

## Gebruiken

1. Installeer MW3 met de Engelse gamebestanden en installeer/update Plutonium
   via de officiële launcher. Start Plutonium eerst eenmaal normaal.
2. Open `SurvivalReimagined.exe`. Kies de eerste keer je MW3-map met `main`
   en `zone/english`. Plutonium wordt standaard gevonden in
   `%LOCALAPPDATA%\Plutonium`; zo nodig verschijnt ook daarvoor een mapkeuze.
3. Bij volgende starts verschijnt een venster met je opgeslagen locaties.
   Kies **Ja** om die te gebruiken, **Nee** om beide installatiemappen opnieuw
   te kiezen, of **Annuleren** om te stoppen. Nieuwe locaties worden pas opgeslagen
   nadat beide mappen geldig zijn. De vorige installatie wordt niet verwijderd.
   De launcher installeert alleen
   gewijzigde ingebouwde modbestanden en start IW5 met Survival op `mp_dome`
   in LAN-modus. Dit is geen openbare server of online accountlogin.

Sluit de game voordat je deze launcher start. Installeer geen tweede mod tegelijk
in dezelfde Plutonium-storage: deze mod gebruikt ook gedeelde scripts/zonebestanden.
De launcher downloadt geen gamebestanden, runt niet als administrator, bewaart
geen wachtwoorden en heeft geen aparte .NET- of Go-installatie nodig.
Windows kan waarschuwen omdat deze eigen executable niet digitaal ondertekend is.

## Instellingen en herstel

Instellingen staan in `%LOCALAPPDATA%\SurvivalReimagined\launcher.json`.
Wijzig `map` naar een geïnstalleerde map, bijvoorbeeld `mp_hardhat`.
Zet `test_menu` op `true` voor het hostgeldmenu; zie `../TESTING.md` voor toetsen.
Verwijder dit JSON-bestand om opnieuw je mappen te kiezen.

Bestanden die de launcher vervangt worden eerst bewaard onder
`%LOCALAPPDATA%\SurvivalReimagined\backups`, met hun oorspronkelijke relatieve
pad en een SHA256-suffix. Bestaande backups worden behouden. Om een bestand te
herstellen: sluit de game, kopieer de gewenste backup terug naar hetzelfde pad
onder je Plutonium `storage\iw5` en verwijder de `.HASH.bak`-suffix.
Nieuw geïnstalleerde bestanden hebben geen eerdere versie; die moet je voor
volledig verwijderen zelf verwijderen. De launcher verwijdert geen andere mods.

Versie 0.1.2 toont direct een diagnosevenster dat openblijft tot je op Enter drukt.
De mapkeuze gebruikt een gewoon Windows-venster, zonder PowerShell.
Alle stappen en fouten worden vanaf het opstarten gelogd; ook wanneer de
mapkeuze of installatie mislukt. Foutmeldingen verschijnen ook in een venster.
Uitvoer van Plutonium staat in
`%LOCALAPPDATA%\SurvivalReimagined\last-launch.log`.

## Bouwen en testen

Met Python 3 en Go >= 1.23: `python3 build.py` vanuit deze map.
Resultaat: `dist/SurvivalReimagined.exe` en `dist/SHA256SUMS.txt`.
Er worden geen externe Go-modules gebruikt. De build test eerst installatie,
backups, herhaalbaarheid, padbeveiliging, instellingen en startargumenten.

De Windows-executable is in Linux cross-gecompileerd; de Windows-mapkeuze en
het daadwerkelijk starten/laden van Plutonium zijn hier niet getest.
De wijzigingen aan GSC en de meegeleverde fastfiles zijn niet opnieuw in-game
gevalideerd. Als zelfs het diagnosevenster en logbestand niet verschijnen, controleer dan
of de ZIP is uitgepakt en of Windows-beveiliging het bestand heeft geblokkeerd.
Schakel beveiligingssoftware niet uit; raadpleeg de precieze melding.

Als jouw Plutonium-versie command-line mod/mapselectie anders
afhandelt, moet de startprocedure na een echte Windows-test worden aangepast.

Startpatroon gebaseerd op de IW5 serverdocumentatie en scripts:
https://plutonium.pw/docs/server/iw5/setting-up-a-server/
https://github.com/xerxes-at/IW5ServerConfig/blob/master/!start_mp_server.bat
Dit bevestigt het bootstrapperpatroon, niet een uitgevoerde clienttest.
