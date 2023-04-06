#!/bin/bash
#set -x
PRO=gamesrv

function USAGE(){
    echo "Usage: $0 [Control]. Such as $0 start"
    # echo -e "\tProcessName: [ QPHallServer | QPGateWay ]"
    echo -e "\tControl: [ start | stop | force_stop | restart | force_restart ]"
    exit 1
}

function START(){
    rm ./scroll_all.log
    nohup ./$PRO &
    sleep 5
    ps -ef | grep -v 'grep' | grep -v 'start' | grep ${PRO}
    if [ $? == 0 ]
    then
        echo "${PRO} Started!"
    else
        echo "Start ${PRO} failed!"
        exit 1
    fi
}

function STOP(){
  rm ./scroll_all.log
	pids=$(ps aux | grep $PRO | grep -v grep | awk -F " " '{print $2}')
	arr=(${pids//'
		'/ })
	#echo $arr
	#echo ${#arr[@]}
	for ((i=0;i<${#arr[@]};i++))
	do
	    kill -s 2 ${arr[$i]} > /dev/null 2>&1
	done
	echo "${PRO} has been stopped!"
}

function FORCE_STOP(){
  rm ./scroll_all.log
    pids=$(ps aux | grep $PRO | grep -v grep | awk -F " " '{print $2}')
	arr=(${pids//'
		'/ })
	for ((i=0;i<${#arr[@]};i++))
	do
	    kill -9 ${arr[$i]} > /dev/null 2>&1
	done
    echo "${PRO} has been shutdown!"
}

function RELOAD() {
  pids=$(ps aux | grep $PRO | grep -v grep | awk -F " " '{print $2}')
	arr=(${pids//'
		'/ })
	for ((i=0;i<${#arr[@]};i++))
	do
	    kill -1 ${arr[$i]} > /dev/null 2>&1
	done
    echo "${PRO} has been RELOADED!"
}

main(){
    case $1 in
    start)
        START ;
    ;;
    stop)
        STOP ;
    ;;
    force_stop)
        FORCE_STOP ;
    ;;
    restart)
        STOP ;
        START ;
    ;;
#    reload)
#        RELOAD ;
#    ;;
    force_restart)
        FORCE_STOP ;
        START ;
    esac
}


if [ -z $1 ]
then
    USAGE
fi


main $1




