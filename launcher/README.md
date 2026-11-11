# Complete launcher v0.3.0

Eén ZIP: pak uit en open SurvivalReimagined.exe. Controleer de installatiemappen
en klik Start Survival. De launcher installeert de volledige officiële resources,
bijbehorende bibliotheken en de aangepaste scripts. Losse herstelpakketten zijn
niet nodig. Bestaande bestanden worden eerst geback-upt. Er wordt nu expliciet
load_dsr survival meegegeven voordat de map wordt gestart.

Deze versie vervangt ook afwijkende bestaande modresources door de gecontroleerde
volledige officiële baseline. De fastfiles in mods/survival en zone worden gelijk
gehouden. Maak bij voorkeur vooraf een backup van je volledige storage/iw5.
Er worden geen MW3-gamebestanden vervangen.

Dit pakket is gebouwd en statisch gecontroleerd, niet op Windows/in-game getest.
De eerdere Defender-melding is niet als foutief bevestigd; omzeil geen blokkade.

Bouwen: Python 3 en Go >=1.23. Pak de originele officiële v4.0.1-release uit;
IW5_OFFICIAL_DIR moet wijzen naar die uitgepakte map. Run python3 launcher/build.py.
Elke baselinefile wordt gecontroleerd tegen launcher/official-files.json.
Bron: https://github.com/LastDemon99/IW5-Survival-Reimagined/releases/tag/iw5-mp-survival-v4.0.1

## Gebruiken

1. Installeer/update Plutonium via de officiële launcher en start het eenmaal.
   Je hebt ook een bestaande Engelse MW3-installatie nodig.
2. Pak het ZIP-pakket uit en open `SurvivalReimagined.exe`.
3. Klik bij **MW3-installatiemap** op **Bladeren** en kies de map met `main`
   en `zone/english`.
4. Controleer de **Plutonium-map**, standaard `%LOCALAPPDATA%\Plutonium`.
   Je kunt die met **Bladeren** wijzigen; de map moet `bin` bevatten.
5. Kies eventueel een andere startmap (standaard `mp_dome`). Het host-testmenu
   is optioneel; zie `TESTING.md` voor de toetsen en beperkingen.
6. **Instellingen opslaan** onthoudt de locaties zonder modbestanden te kopiëren.
   **Start Survival** slaat de instellingen op, controleert/bijwerkt de mod en
   start de bestaande Plutonium-bootstrapper in LAN-modus.

Bij een volgende start staan de opgeslagen locaties in hetzelfde venster. Je
kunt ze altijd weer wijzigen. Je hoeft niets in een opdrachtprompt te typen.
Tijdens installeren en zolang Plutonium draait zijn wijzigingen geblokkeerd.
Na afsluiten van Plutonium wordt de startknop weer beschikbaar.

De status onderin toont de voortgang. Bij fouten verschijnt een melding.
Met **Log bekijken** open je het logbestand in Windows Kladblok.

## Je bestaande installatie

Maak eerst zelf een backup van je volledige Plutonium `storage\iw5`-map.
De launcher wijzigt geen MW3-gamebestanden. Hij plaatst de modbestanden in
`<Plutonium-map>\storage\iw5`; ook gedeelde scripts en zonebestanden kunnen
worden vervangen. Identieke bestanden worden overgeslagen.

Van vervangen bestanden wordt eerst een backup gemaakt onder
`%LOCALAPPDATA%\SurvivalReimagined\backups`. Bestandsnamen bevatten een
SHA256-suffix. Sluit de game voordat je bestanden herstelt; kopieer een backup
naar het oorspronkelijke relatieve pad onder `storage\iw5` en verwijder de
`.HASH.bak`-suffix. Nieuwe bestanden hebben geen eerdere versie om te herstellen.
De launcher verwijdert geen andere bestanden en geen oude installatie wanneer
je een andere map kiest. Start geen twee launchers tegelijk.

Instellingen: `%LOCALAPPDATA%\SurvivalReimagined\launcher.json`.
Log: `%LOCALAPPDATA%\SurvivalReimagined\last-launch.log`.

## Bouwen en valideren

Vanuit deze map: `python3 build.py`, met Python 3 en Go >= 1.23.
Dit voert de zes installatietests uit, bouwt de Windows-executable
 en maakt `dist/SurvivalReimagined-Windows-v0.3.0.zip` met checksums en instructies.
Geen externe Go-modules. Het programma heeft Windows' GUI-subsystem en gebruikt
standaard Win32-controls; debug-symbolen worden niet gestript. VCS-buildmetadata
wordt weggelaten voor reproduceerbare builds tussen schone checkouts.

Gecontroleerd in Linux: zes tests (inclusief installatie van de originele herstelafbeelding), Windows cross-compilatie, Windows
`go vet`, GUI-subsystem en ZIP-integriteit. Niet getest: werkelijk tekenen en
bedienen van het venster, Windows-mapkeuze, Plutonium-start en gameplay. Hiervoor
ontbreken hier Windows/Wine en MW3. Een geslaagde build is geen antivirusvrijgave.

Windows-handmatige controle: wijzig beide mappen, annuleer Bladeren zonder
veldinformatie te verliezen, sla op en heropen, controleer ongeldige paden en
mapnamen, start een match, open het log en controleer herstel na een startfout.

Startpatroon gebaseerd op:
https://plutonium.pw/docs/server/iw5/setting-up-a-server/
https://github.com/xerxes-at/IW5ServerConfig/blob/master/!start_mp_server.bat
