#!/bin/bash
if [[ debug == "$1" ]]; then
  INSTRUMENTING=yes  # any non-null will do
  shift
fi
debugecho () {
  [[ "$INSTRUMENTING" ]] && builtin echo $@
}

[[ "$INSTRUMENTING" ]] && git fetch --tags
[[ "$INSTRUMENTING" ]] && git pull --tags
[[ "$INSTRUMENTING" ]] || git pull --quiet --tags

CURRENT_VER=$(git describe --tags --abbrev=0)
CURRENT_HASH=$(git rev-parse --short HEAD)
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)

if [[ $CURRENT_BRANCH != "develop" && $CURRENT_BRANCH != "staging" ]]
then
    debugecho "failed branch check: " $CURRENT_BRANCH
    echo invalid branch, should be develop or staging: current is $CURRENT_BRANCH
    exit 1
fi

case $CURRENT_VER in
  *rc.*)
    debugecho "case *rc.*" 
    NEXT_VER=$(echo $CURRENT_VER | awk -F. -v OFS=. 'NF==1{print ++$NF}; NF>1{if(length($NF+1)>length($NF))$(NF-1)++; $NF=sprintf("%0*d", length($NF), ($NF+1)%(10^length($NF))); print}')
    ;;

  *-dev*)
    debugecho "case *-dev*" 
    STRIPED_VER=$(echo $CURRENT_VER | cut -f1 -d"-")
    NEXT_VER=$(printf "%s-dev-%s" $STRIPED_VER $CURRENT_HASH)
    ;;

  *)
    debugecho "case \*"
    if [[ $CURRENT_BRANCH == "develop" ]];
    then
        debugecho "case \* if develop"
        NEXT_VER=$(printf "%s-dev-%s" $STRIPED_VER $CURRENT_HASH)
    elif [[ $CURRENT_BRANCH == "staging" ]];
    then
        debugecho "case \* if staging"
        NEXT_VER=$(printf "%s-rc.1" $CURRENT_VER)
    else
        debugecho "case \* else"
        NEXT_VER=$(printf "%s-rc.1" $CURRENT_VER)
        echo invalid branch, should be develop or staging
        exit 1
    fi
    ;;

esac

debugecho $CURRENT_VER "->" $NEXT_VER
printf "%s\n" $NEXT_VER
