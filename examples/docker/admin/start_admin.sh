mkdir -p $PONZU_SHARE/uploads
mkdir -p $PONZU_SHARE/search
touch $PONZU_SHARE/system.db
touch $PONZU_SHARE/analytics.db
ln -sf $PONZU_SHARE/uploads $PROJECT_FOLDER/uploads
ln -sf $PONZU_SHARE/search $PROJECT_FOLDER/search
ln -sf $PONZU_SHARE/system.db $PROJECT_FOLDER/system.db
ln -sf $PONZU_SHARE/analytics.db $PROJECT_FOLDER/analytics.db

# ponzu build",
# ponzu -port=8080 --https run admin,api"