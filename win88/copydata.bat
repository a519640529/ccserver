@echo off
xcopy R:\gocode\trunk\src\games.agamestudio.com\jxjyhj\data\*.dat R:\quick-jxjyhj\trunk\jxjyqp\res\data /s /e /y /Q
xcopy R:\gocode\trunk\src\games.agamestudio.com\jxjyhj\protocol_lua\*.proto R:\quick-jxjyhj\trunk\jxjyqp\protocol /s /e /y /Q
if exist "R:\quick-jxjyhj\trunk\jxjyqp\src\protocol\pbdata_pb.lua" (del R:\quick-jxjyhj\trunk\jxjyqp\src\protocol\pbdata_pb.lua)
if exist "R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_NameBoy.dat" (del R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_NameBoy.dat)
if exist "R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_NameGirl.dat" (del R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_NameGirl.dat)
if exist "R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_Sensitive_Words.dat" (del R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_Sensitive_Words.dat)
if exist "R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_FishPool.dat" (del R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_FishPool.dat)
if exist "R:\quick-jxjyhj\trunk\jxjyqp\src\protocol\db_fishpool_pb.lua" (del R:\quick-jxjyhj\trunk\jxjyqp\src\protocol\db_fishpool_pb.lua)
if exist "R:\quick-jxjyhj\trunk\jxjyqp\protocol\db_fishpool.proto" (del R:\quick-jxjyhj\trunk\jxjyqp\protocol\db_fishpool.proto)
if exist "R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_SlotsPool.dat" (del R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_SlotsPool.dat)
if exist "R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_DNPlayerOP.dat" (del R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_DNPlayerOP.dat)
if exist "R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_RobotChangeCard.dat" (del R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_RobotChangeCard.dat)
if exist "R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_ZJHCardKindValue.dat" (del R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_ZJHCardKindValue.dat)
if exist "R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_ZJHPlayerOP.dat" (del R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_ZJHPlayerOP.dat)
if exist "R:\quick-jxjyhj\trunk\jxjyqp\res\data\LB_PokerLibrary.dat" (del R:\quick-jxjyhj\trunk\jxjyqp\res\data\LB_PokerLibrary.dat)
if exist "R:\quick-jxjyhj\trunk\jxjyqp\src\protocol\db_slotspool_pb.lua" (del R:\quick-jxjyhj\trunk\jxjyqp\src\protocol\db_slotspool_pb.lua)
if exist "R:\quick-jxjyhj\trunk\jxjyqp\protocol\db_slotspool.proto" (del R:\quick-jxjyhj\trunk\jxjyqp\protocol\db_slotspool.proto)




if exist "R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_WinThreeAI_Weight.dat" (del R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_WinThreeAI_Weight.dat)
if exist "R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_WinThreeAI_CardKind.dat" (del R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_WinThreeAI_CardKind.dat)
if exist "R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_TuiTongPlayerOP.dat" (del R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_TuiTongPlayerOP.dat)
if exist "R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_TuiTongCardKindValue.dat" (del R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_TuiTongCardKindValue.dat)
if exist "R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_PUBG_WeightCondition.dat" (del R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_PUBG_WeightCondition.dat)
if exist "R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_PUBG_Weight.dat" (del R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_PUBG_Weight.dat)
if exist "R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_PUBG_TurnRate.dat" (del R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_PUBG_TurnRate.dat)
if exist "R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_PUBG_Odds.dat" (del R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_PUBG_Odds.dat)
if exist "R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_JDDNCardKind_NewAI.dat" (del R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_JDDNCardKind_NewAI.dat)
if exist "R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_FortuneGod_WeightCondition.dat" (del R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_FortuneGod_WeightCondition.dat)
if exist "R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_FortuneGod_Weight.dat" (del R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_FortuneGod_Weight.dat)
if exist "R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_FortuneGod_TurnRate.dat" (del R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_FortuneGod_TurnRate.dat)
if exist "R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_FortuneGod_Odds.dat" (del R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_FortuneGod_Odds.dat)
if exist "R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_DZPKPlayerOP.dat" (del R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_DZPKPlayerOP.dat)
if exist "R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_DZPKCardKindValue.dat" (del R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_DZPKCardKindValue.dat)
if exist "R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_DragonVsTigerPool.dat" (del R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_DragonVsTigerPool.dat)
if exist "R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_DNCardKind_NewAI.dat" (del R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_DNCardKind_NewAI.dat)
if exist "R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_DNBullCardKind.dat" (del R:\quick-jxjyhj\trunk\jxjyqp\res\data\DB_DNBullCardKind.dat)

pause