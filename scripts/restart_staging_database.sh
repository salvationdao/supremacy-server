read -p "Are you sure you want to reset staging data on Gameserver? (y/n)" -n 1 -r yn
case "$yn" in
    [yY] )  echo ""
            echo "Proceeding to reset data"
            ;;
    [nN] )  echo ""
            echo "Exiting.."
            exit
            ;;
    * )     echo "Invalid response...exiting"
            exit
            ;;
esac
read -p "Have you stopped any related services? (y/n)" -n 1 -r yn
case "$yn" in
    [yY] )  echo ""
            echo "Proceeding to reset data"
            ;;
    [nN] )  echo ""
            echo "Exiting.."
            exit
            ;;
    * )     echo "Invalid response...exiting"
            exit
            ;;
esac
echo ""
echo "What comp number is this (1,2,3, etc) : "
read comp
date=$(date +'%Y-%m-%d-%H%M%S')
psql -U postgres -c "ALTER DATABASE \"gameserver-db\" RENAME TO \"gameserver-db-COMP$comp-$date\""
psql -U postgres -c "CREATE DATABASE \"gameserver-db\" WITH TEMPLATE \"gameserver-db-POST-SETUP-2022-08-05\""
psql -U postgres -d "gameserver-db" -c "ALTER DEFAULT PRIVILEGES IN SCHEMA \"public\" GRANT SELECT ON TABLES TO \"dev-readonly\";"
psql -U postgres -d "gameserver-db" -c "ALTER DEFAULT PRIVILEGES IN SCHEMA \"public\" GRANT ALL PRIVILEGES ON TABLES TO \"gameserver\";"
echo "DONE RESETTING GAMESERVER"
