#deploy server

gcloud builds submit --substitutions=_BUILD_TARGET=server,_SERVICE_NAME=tacoserver,_RUN_HASH=$cloudhash


gcloud builds submit --substitutions=_BUILD_TARGET=client,_SERVICE_NAME=tacoclient,_RUN_HASH=$cloudhash