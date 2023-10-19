class CD_Utils extends Info;

public static function int GetMaxMonsters(WorldInfo WI) {
    local string Output;

    Output = WI.ConsoleCommand("GetAll CD_Survival MaxMonstersInt", false);
    Output = Split(Output, "= ", true);

    if (Output == "") return 0;

    return int(Output);
}

public static function int GetWaveSizeFakes(WorldInfo WI) {
    local string Output;

    Output = WI.ConsoleCommand("GetAll CD_Survival WaveSizeFakesInt", false);
    Output = Split(Output, "= ", true);

    if (Output == "") return 0;

    return int(Output);
}

public static function string GetSpawnCycle(WorldInfo WI) {
    local string Output;

    Output = WI.ConsoleCommand("GetAll CD_Survival SpawnCycle", false);
    Output = Split(Output, "= ", true);

    return Output;
}

public static function string GetZedsType(WorldInfo WI) {
    local string Output;

    Output = WI.ConsoleCommand("GetAll CD_Survival ZedsType", false);
    Output = Split(Output, "= ", true);

    return Output;
}