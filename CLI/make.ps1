$DATE=Get-Date -Format "yyy.mm.dd//HH:mm:ss"
$GITHASH=git rev-parse HEAD
$GITBRANCH=git branch --show-current
$GITHASHDATE=git show -s --format=%ci HEAD | %{$_ -replace " ", "/"}
$DIR=(get-location).path
docker run --rm -v ${DIR}:/workdir -w /workdir -e GOOS=windows golang go build -ldflags="`
    -X cli/controllers.BuildHash=${GITHASH}`
    -X cli/controllers.BuildTree=${GITBRANCH}`
    -X cli/controllers.BuildTime=${$DATE}`
    -X cli/controllers.GitCommitDate=${GITHASHDATE}`
" -o cli.exe