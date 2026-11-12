# Videoprofiel v0.5.1

Bij een gekozen video-optie zoekt de launcher nu vóór de gamestart naar een
bestaande config_mp.cfg. Volgorde: Plutonium/storage/iw5/players, daarna players2;
vervolgens de players2- en players-map in je geselecteerde MW3-installatie.
Alleen de geselecteerde videowaarden worden gewijzigd. Andere instellingen,
naam en binds blijven behouden. Er wordt eerst een exacte backup gemaakt in
%LOCALAPPDATA%/SurvivalReimagined/backups/video-profile. Het geselecteerde pad
staat in Log bekijken.

Als alles op Bestaande instelling staat, wordt geen videoprofiel aangeraakt.
Een ontbrekend, alleen-lezen of onbekend gecodeerd profiel geeft een melding;
er wordt geen configuratie op een gegokte locatie aangemaakt. Start MW3 eenmaal
via de officiële launcher als er nog geen profiel bestaat. De launcher weigert
updates als een Plutonium/MW3-game nog draait of een andere launcher al start.

Dit is de nieuwe route voor systemen waar de losse startopties geen effect
hadden. Bestaande gekozen video-opties worden automatisch hergebruikt. De tests
controleren backups, herhaalbaarheid, BOM/UTF-16/regelafbrekingen, behoud van binds
en foutgevallen. Op Windows worden bovendien de native startvergrendeling en
profielschrijftests uitgevoerd. Het daadwerkelijke MW3-renderresultaat moet nog
op de gebruikersinstallatie worden bevestigd.

# Video-opties v0.5.0

Open VIDEO in de linker navigatie. Kies Venstermodus of Volledig scherm en
optioneel een resolutie, VSync en FPS-limiet. Standaard blijft alles op Bestaande
instelling: oude launcherconfiguraties geven geen nieuwe video-overrides mee.

VENSTER VOOR SCREENSHOTS selecteert venstermodus en 1280x720. Klik VIDEO OPSLAAN
en start daarna Survival. De instellingen gelden bij de volgende gamestart;
sluit een lopende game eerst af. Je hoeft geen mod te unloaden of handmatig
configuratiebestanden te bewerken. De launcher schrijft alleen zijn eigen JSON
instellingen en geeft de keuzes als gecontroleerde startup-dvars door:
r_fullscreen, r_mode, r_vsync en com_maxfps.

De backendtests controleren defaults, validatie, volgorde vóór het mapcommando
en het opslaan/herlezen van alle video-opties. De Windows-smoketest opent VIDEO,
controleert het screenshot-preset en maakt een echte screenshot van het venster.
De daadwerkelijke MW3-rendering en welke resoluties een monitor ondersteunt zijn
niet met deze smoketest getest. Kies alleen een passende schermresolutie.

# Testmenu en scriptprioriteit v0.4.3

De project-GSC-scripts worden tevens in zz_survival_reimagined.iwd geplaatst,
zowel in storage/iw5 als mods/survival. De zz-naam volgt de aflopende archief-
zoekvolgorde die het gebruikerslog toont. Dit laat de aangepaste scripts vóór
de oorspronkelijke LB_Survival-scripts vinden. De originele baselinearchieven
blijven behouden. De compiler moet nu ook dev/testmenu.gsc in console.log noemen.

Schakel het geldtestmenu in bij Instellingen. Gebruik daarna de bestaande toets
of controllerknop voor actieslot 5; F5 is NIET verplicht. Actieslot 6 geeft $1.000,
actieslot 7 geeft $10.000 wanneer het paneel open is. Als die acties nog geen
toets hebben, moeten ze wel in de gamebediening worden gekoppeld. Bestaande
bindings worden niet overschreven. De HUD verwijst naar de ingestelde binding.
Gebruik een privétest zonder streaks die dezelfde actieslots gebruiken.

De levering, hashgelijkheid van scripts en UI worden getest; echte GSC-compilatie
en in-game menubediening zijn nog niet op deze build uitgevoerd.

# LAN-spelernaam v0.4.2

Vul in INSTELLINGEN de spelernaam (LAN) in. Deze wordt onthouden en expliciet als
+name meegegeven. De launcher gebruikt de extra opstartopties uit de officiële
Plutonium-launcher niet, omdat hij de bootstrapper rechtstreeks start.
Je forumaccount wordt niet gewijzigd. Bestaande configuraties krijgen Survivor
als neutrale standaardnaam. Je kunt die vóór het starten aanpassen.
Een naam van 16 tekens wordt door de launcher ongewijzigd doorgegeven.
Dit herstelt het concrete verschil in startargumenten; een native crash met
Crash reference ID 0 is niet op basis van alleen die melding verklaard.

# Opstartresources v0.4.2

De originele mod.iwd wordt bij installatie tevens als LB_Survival_Resources.iwd
in de actieve storage/iw5-map geplaatst. Zo zijn afbeeldingen en geluiden beschikbaar
wanneer de root-zone fastfiles laden, vóór mods/survival gemount is. Het zijn exact
dezelfde originele bytes, geen vervangende textures. Beide archieven krijgen de
bestaande backup- en herstelbehandeling. Het resourcearchief staat in gedeelde
storage en is daardoor ook buiten een Survival-match vindbaar.

Deze verandering volgt het gebruikerslog: de bibliotheken waren gemount, maar de
modresource-IWD ontbrak uit het vroege zoekpad bij de weapon_famas-fout.
De installatietest controleert beide locaties en de drie eerder genoemde images.
De volledige gameplay moet nog op de gebruikersinstallatie worden bevestigd.

# Survival Reimagined Launcher v0.4.2 — Operations UI

Donkere, game-geïnspireerde vormgeving met olijfgroene accenten en originele
vectorillustratie van een tactische kaart. SPELEN toont de startknop en startmap;
INSTELLINGEN bevat de installatiemappen en host-testoptie. LOG BEKIJKEN opent het
bestaande log. De volledige modresources en back-uplogica uit v0.3.0 blijven gelijk.
Alle UI-tekst en illustraties worden lokaal getekend: geen webview, netwerkassets,
externe lettertypen of aanvullende runtime vereist.

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
Dit voert de dertien installatietests uit, bouwt de Windows-executable
 en maakt `dist/SurvivalReimagined-Windows-v0.4.2.zip` met checksums en instructies.
Geen externe Go-modules. Het programma heeft Windows' GUI-subsystem en gebruikt
standaard Win32-controls; debug-symbolen worden niet gestript. VCS-buildmetadata
wordt weggelaten voor reproduceerbare builds tussen schone checkouts.

Gecontroleerd in Linux: dertien tests, Windows cross-compilatie, Windows `go vet`,
GUI-subsystem en ZIP-integriteit. De release-workflow start de echte executable
op Windows, controleert de speel- en instellingenpagina en het sluiten van de
app, en maakt screenshots. Publiceren gebeurt pas na deze Windows-smoketest.
Windows-mapkeuze, Plutonium-start en gameplay vallen buiten deze smoketest.

Windows-handmatige controle: wijzig beide mappen, annuleer Bladeren zonder
veldinformatie te verliezen, sla op en heropen, controleer ongeldige paden en
mapnamen, start een match, open het log en controleer herstel na een startfout.

Startpatroon gebaseerd op:
https://plutonium.pw/docs/server/iw5/setting-up-a-server/
https://github.com/xerxes-at/IW5ServerConfig/blob/master/!start_mp_server.bat
