xcopy .\data R:\gocode\ReadyUpdate\data /s /e /y
xcopy .\datasrv\datasrv R:\gocode\ReadyUpdate /y
del .\datasrv\datasrv
xcopy .\gatesrv\gatesrv R:\gocode\ReadyUpdate /y
del .\gatesrv\gatesrv
xcopy .\gamesrv\gamesrv R:\gocode\ReadyUpdate /y
del .\gamesrv\gamesrv
xcopy .\mgrsrv\mgrsrv R:\gocode\ReadyUpdate /y
del .\mgrsrv\mgrsrv
xcopy .\robot\robot R:\gocode\ReadyUpdate /y
del .\robot\robot
xcopy .\routesrv\routesrv R:\gocode\ReadyUpdate /y
del .\routesrv\routesrv
xcopy .\worldsrv\worldsrv R:\gocode\ReadyUpdate /y
del .\worldsrv\worldsrv
xcopy .\schedulesrv\schedulesrv R:\gocode\ReadyUpdate /y
del .\schedulesrv\schedulesrv


if exist "R:\gocode\ReadyUpdate\data\gameparam.json" (del R:\gocode\ReadyUpdate\data\gameparam.json)
if exist "R:\gocode\ReadyUpdate\data\thrconfig.json" (del R:\gocode\ReadyUpdate\data\thrconfig.json)
if exist "R:\gocode\ReadyUpdate\data\LB_PokerLibrary.json" (del R:\gocode\ReadyUpdate\data\LB_PokerLibrary.json)
if exist "R:\gocode\ReadyUpdate\data\LB_PokerLibrary.dat" (del R:\gocode\ReadyUpdate\data\LB_PokerLibrary.dat)
if exist "R:\gocode\ReadyUpdate\data\gamedata.json" (del R:\gocode\ReadyUpdate\data\gamedata.json)
if exist "R:\gocode\ReadyUpdate\data\gmac.json" (del R:\gocode\ReadyUpdate\data\gmac.json)
if exist "R:\gocode\ReadyUpdate\data\zone_rob.json" (del R:\gocode\ReadyUpdate\data\zone_rob.json)
if exist "R:\gocode\ReadyUpdate\data\icon_rob.json" (del R:\gocode\ReadyUpdate\data\icon_rob.json)
pause