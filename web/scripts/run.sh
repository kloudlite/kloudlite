if [[ -z $APP ]]; then
	echo 'ERROR: set APP env var'
	exit 1
fi

remix-serve ./public/${APP}/server
