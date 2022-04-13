require 'sinatra'

Dir.chdir('/usr/share/ninja_syndicate')

`download_version.sh ${GAME_VERSION}`
`echo 'y' | ./change_version.sh gameserver_${GAME_VERSION}`

