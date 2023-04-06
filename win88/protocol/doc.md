#协议id段划分

--------------------------------------
- 1000~1999 server->server
- 2000~4999 client->worldsrv 
- 5000~9999 client->gamesrv
--------------------------------------

##worldsrv

###login（登录相关）
- 2001~2099

###player（玩家相关）
- 2100~2199

###gamehall
####game.proto
- 2200-2319

####coinscene.proto
- 2320-2339

####hallpacket.proto
- 2340-2379

####hundredscene.proto
- 2380-2399

####task
- 2400~2429

####message （邮箱，通知）
- 2430~2449

####match （比赛）
- 2450~2499

####shop（商店）
- 2500~2529

####bag（背包）
- 2530~2549

####Pets（人物宠物）
- 2550~2579

####Welfare（福利）
- 2580~2599

####activity（活动相关）
- 2600~2699

####friend（好友）
- 2700~2719

####chat（聊天）
- 2720~2739

####tournament（锦标赛）
- 2740~2759

##gamesrv

###fish
- 5000~5099

###fortunegod
- 5100~5119

###baccarat
- 5120~5139

###crash
- 5140~5159

###fortunezhishen
- 5160~5179

###avengers
- 5180~5199

###easterisland
- 5200~5219

###caishen（财神）
- 5220~5239

###糖果 candy.proto
- 5240~5259

###caothap.proto
- 5260~5279

###luckydice.proto
- 5280~5299

###rollcoin.proto
- 5300~5319

###blackjack.proto
- 5320~5339

###dezhoupoker.proto
- 5340~5369

###tienlen.proto
- 5370~5389

###hundredyxx.proto
- 5390~5409

###roulette.proto
- 5410~5429

###rollpoint.proto
- 5430~5449
- 
###dragonvstiger.proto
- 5450~5469
- 
###redvsblack.proto
- 5470~5489
-
###fruits.proto
- 5490~5499
-
###richblessed.proto
- 5500~5509
- 
###game.proto(玩家离开)
- 8000~8099