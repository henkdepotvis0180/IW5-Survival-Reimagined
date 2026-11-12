#include lethalbeats\survival\utility;
#include lethalbeats\hud;

watch()
{
	level endon("game_ended");
	self endon("disconnect");
	self endon("death");
	if (self isTestClient() || !self isHost()) return;

	self notifyOnPlayerCommand("survival_test_toggle", "+actionslot 5");
	self notifyOnPlayerCommand("survival_test_small", "+actionslot 6");
	self notifyOnPlayerCommand("survival_test_large", "+actionslot 7");
	if (getDvarInt("survival_test_menu"))
		self iPrintLn("Testmenu: druk op [{+actionslot 5}] om te openen.");
	panel = hud_create_string(self, "TEST CASH | [{+actionslot 6}]: +$1,000 | [{+actionslot 7}]: +$10,000 | [{+actionslot 5}]: sluiten", "hudsmall", 0.7, "TOP CENTER", "TOP CENTER", 0, 65);
	panel.alpha = 0;
	self thread cleanup(panel);
	opened = false;
	for (;;)
	{
		event = self lethalbeats\utility::waittill_any_return("survival_test_toggle", "survival_test_small", "survival_test_large");
		// Recheck server permission for every action, including changes mid-match.
		if (!getDvarInt("survival_test_menu") || !self isHost() || !isAlive(self) || self.inLastStand)
		{
			opened = false;
			panel.alpha = 0;
			continue;
		}
		if (event == "survival_test_toggle")
		{
			opened = !opened;
			panel.alpha = int(opened);
			continue;
		}
		if (!opened || !isDefined(self.pers["score"])) continue;
		amount = event == "survival_test_small" ? 1000 : 10000;
		// Test cash updates the shop/HUD balance without inflating earned-money stats.
		balance = self.pers["score"] + amount;
		if (balance > 1000000) balance = 1000000;
		self survivor_set_score(balance);
		self iPrintLn("Test balance: $" + balance);
		wait 0.2;
	}
}

cleanup(panel)
{
	self lethalbeats\utility::waittill_any("death", "disconnect");
	if (isDefined(panel)) panel destroy();
}
