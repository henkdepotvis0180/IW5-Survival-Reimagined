# Plutonium checks

These source changes require in-game validation. No fastfiles were rebuilt.
Back up your existing Plutonium storage before copying the changed loose GSC
files from `iw5/` into the corresponding storage paths, including the new
`lethalbeats/Survival/dev/testMenu.gsc`. Restart the match after installation.

## Graphics

The intro now uses an invisible camera origin instead of a visible cinematic
rig, with cleanup on death/disconnect. This is a candidate fix, not a confirmed
diagnosis of stretched geometry. Compare the same map with
`set survival_intro_camera 0` before spawning to bypass the intro entirely.
Check normal spawn, death during the intro, respawn, dog attacks and juggernaut
helicopter drops. Confirm HUD, camera, controls and ADS recover. If geometry
still stretches, record the map, wave, enemy type and console errors; inspect
the custom models/animations and installed fastfile versions next.

## AI

Easy initial/reacquisition reaction delays change from 0.35s to 0.25s; normal
changes from 0.15/0.12s to 0.12/0.10s. Aim-bone refresh changes from 0.40s to
0.25s on easy and 0.25s to 0.15s on normal. Hard, damage and accuracy settings
are unchanged. Existing `bot_skill_*` overrides still take precedence.
Compare target acquisition and moving-target tracking on easy/normal; test
stuns, last stand, and navigation separately. The botactor implementation is
not supplied as loose source here, so targeting/pathfinding is not validated.

## Host test cash panel

Enable with `set survival_test_menu 1` in the host/server console. Bind unused
keys to `+actionslot 5`, `+actionslot 6`, and `+actionslot 7` in game settings or
the console. Slot 5 toggles the panel; while open, slot 6 adds $1,000 and slot 7
adds $10,000. These slots can also activate existing game actions: use a private
test match without equipped streaks and do not combine with the map editor.
Only the listen-server host can use it; dedicated-server remote admin access
is not implemented. Balance is capped at $1,000,000 and earned-money stats are
not increased. Disable with `set survival_test_menu 0`.

Check disabled mode, non-host access, both cash buttons, purchases after giving
cash, balance cap, closing, death/respawn and disconnect. Disabling the setting
must reject further grants, even if the panel was previously open.

## Checks performed in the cloud

`git diff --check` and delimiter checks on modified scripts passed. These are
static checks, not GSC compilation or Plutonium gameplay tests. This Linux
environment lacks the required game runtime and MW3 assets.
