if [[ -z $APP ]]; then
	echo 'ERROR: set APP env var'
	exit 1
fi

remix build
if [ $? -eq 0 ]; then
	pushd public/${APP}/server > /dev/null 2>&1
	if [ ! -f package.json ]; then
		pnpm init > /dev/null 2>&1
	fi
	popd > /dev/null 2>&1
fi
