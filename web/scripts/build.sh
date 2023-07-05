if [[ -z $APP ]]; then
	echo 'ERROR: set APP env var'
	exit 1
fi


remix build
echo {\"type\":\"commonjs\"} > public/${APP}/server/package.json