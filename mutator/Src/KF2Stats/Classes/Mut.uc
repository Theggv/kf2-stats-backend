class Mut extends KFMutator;

var private KF2Stats KF2Stats;

var float LastZedTimeTimestamp;

public simulated function bool SafeDestroy()
{
	return (bPendingDelete || bDeleteMe || Destroy());
}

public function ScoreKill(Controller Killer, Controller Other) {
	local KFPlayerController C;
	local name Zedkey;

	C = KFPlayerController(Killer);

	if (Killer == None || Killer.PlayerReplicationInfo == None || 
		Killer == Other || Other == None || Other.Pawn == None) {
		return;
	} 

	if (Other.Pawn.IsA('KFPawn_Monster')) {
		Zedkey = KFPawn_Monster(Other.Pawn).LocalizationKey;
		KF2Stats.Stats.AddZedKill(C, Zedkey);
	}
}

function NetDamage(
	int OriginalDamage,
	out int Damage, 
	Pawn Injured, 
	Controller InstigatedBy, 
	vector HitLocation, 
	out vector Momentum, 
	class<DamageType> DamageType, 
	Actor DamageCauser
) {
	local name ZedKey;

	super.NetDamage(OriginalDamage, Damage, Injured, InstigatedBy, HitLocation, Momentum, DamageType, DamageCauser);

	// `log(
	// 	"instigator="$InstigatedBy @
	// 	"injured="$Injured @
	// 	"orig_dmg="$OriginalDamage @
	// 	"mod_dmg="$Damage @
	// 	"type="$DamageType @
	// 	"causer="$DamageCauser
	// );

	// Detect fleshpound rage from husk backback
	if (KFPlayerController(InstigatedBy) != None &&
		KFPawn_ZedFleshpound(Injured) != None &&
		DamageType == class'KFDT_Explosive_HuskSuicide'
	) {
		DetectFPRageFromHuskBP(Damage, KFPawn_ZedFleshpound(Injured), KFPlayerController(InstigatedBy));
	}

	// Detect husk backback kill
	if (KFPlayerController(InstigatedBy) != None &&
		KFPawn_ZedHusk(Injured) != None && Damage == 10000
	) {
		KF2Stats.Stats.AddEvent(KFPlayerController(InstigatedBy), ET_HUSK_BACKPACK);
	}

	// Detect damage to player
	if (KFPawn_Human(Injured) != None && KFPawn_Monster(InstigatedBy.Pawn) != None) {
		Zedkey = KFPawn_Monster(InstigatedBy.Pawn).LocalizationKey;
		KF2Stats.Stats.AddInjuredByZed(KFPlayerController(Injured.Controller), Zedkey, Damage);
	}
}

// detection method from phanta's cd chokepoints
private function DetectFPRageFromHuskBP(
	int Damage, 
	KFPawn_ZedFleshpound Injured, 
	KFPlayerController InstigatedBy
) {
	local KFAIController_ZedFleshpound AI;
	local KFAIPluginRage_Fleshpound RagePlugin;
	local DamageModifierInfo DamageModifier;
	local float mp;

	AI = KFAIController_ZedFleshpound(Injured.Controller);
	if (AI == None) return;

	RagePlugin = AI.RagePlugin;

	if (RagePlugin == None) return;
	if (RagePlugin.bIsEnraged) return;

	mp = 1.0;
	foreach class'KFPawn_ZedFleshpound'.default.DamageTypeModifiers(DamageModifier) {
		if (DamageModifier.DamageType != class'KFDT_Explosive') continue;

		mp = DamageModifier.DamageScale[0];
	}

	// `log(
	// 	"multiplier="$mp @
	// 	"dmg="$RagePlugin.AccumulatedDOT + float(Damage) * mp @
	// 	"threshold="$RagePlugin.RageDamageThreshold
	// );

	if (float(RagePlugin.AccumulatedDOT) + float(Damage) * mp < RagePlugin.RageDamageThreshold) return;

	// `log("Detected fp rage from bp");
	KF2Stats.Stats.AddEvent(InstigatedBy, ET_RAGED_BY_BP);
}

function ModifyZedTime(
	out float out_TimeSinceLastEvent, 
	out float out_ZedTimeChance, 
	out float out_Duration
) {
	super.ModifyZedTime(out_TimeSinceLastEvent, out_ZedTimeChance, out_Duration);

	// no idea how to detect initial zedtime
	if (out_ZedTimeChance >= 1.0) {
		if (WorldInfo.RealTimeSeconds - LastZedTimeTimestamp > 3.0) {
			LastZedTimeTimestamp = WorldInfo.RealTimeSeconds - 5.8; 
			KF2Stats.Stats.SessionData.ZedTimeCount += 1;
		}

		KF2Stats.Stats.SessionData.ZedTimeDuration += (out_Duration - 
			(out_Duration - (WorldInfo.RealTimeSeconds - LastZedTimeTimestamp)));

		LastZedTimeTimestamp = WorldInfo.RealTimeSeconds;
	}
}

public event PreBeginPlay() {
	Super.PreBeginPlay();

	if (WorldInfo.NetMode == NM_Client) return;

	foreach WorldInfo.DynamicActors(class'KF2Stats', KF2Stats) {
		break;
	}

	if (KF2Stats == None) {
		KF2Stats = WorldInfo.Spawn(class'KF2Stats');
	}

	if (KF2Stats == None)
	{
		`Log_Base("FATAL: Can't Spawn 'KF2Stats'");
		SafeDestroy();
	}
}

public function AddMutator(Mutator Mut)
{
	if (Mut == Self) return;

	if (Mut.Class == Class)
		Mut(Mut).SafeDestroy();
	else
		Super.AddMutator(Mut);
}

defaultproperties
{

}