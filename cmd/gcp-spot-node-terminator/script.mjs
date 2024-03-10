#! /usr/bin/env zx

$.verbose = false

const nodename = process.env.NODE_NAME
if (!nodename) {
  throw new Error("env var NODE_NAME is not set, exiting.")
}

// const debug = process.env.DEBUG == true

function debug(msg) {
  if (process.env.DEBUG == "true") {
    console.log(`[#] DEBUG ${msg}`)
  }
}

function info(msg) {
  console.log(`[#] INFO ${msg}`)
}

console.log(`
                    ,                       
                  #####                 
               ########                 
             ########                   
          ########        #####             
        ########       *#########            ██╗  ██╗██╗      ██████╗ ██╗   ██╗██████╗ ██╗     ██╗████████╗███████╗
     ########        ###############         ██║ ██╔╝██║     ██╔═══██╗██║   ██║██╔══██╗██║     ██║╚══██╔══╝██╔════╝
   ########       *###################       █████╔╝ ██║     ██║   ██║██║   ██║██║  ██║██║     ██║   ██║   █████╗  
 #######/       ########################     ██╔═██╗ ██║     ██║   ██║██║   ██║██║  ██║██║     ██║   ██║   ██╔══╝  
   #######(        ###################       ██║  ██╗███████╗╚██████╔╝╚██████╔╝██████╔╝███████╗██║   ██║   ███████╗
     (#######.       ##############*         ╚═╝  ╚═╝╚══════╝ ╚═════╝  ╚═════╝ ╚═════╝ ╚══════╝╚═╝   ╚═╝   ╚══════╝
        ########        #(#######                              
          (#######.       ####*               __   ___       __
             ########                        |__) |__   /\  |  \ \ / 
               /######(.                     |  \ |___ /~~\ |__/  |  
                  #####
                    ,                       

`)

const metadataHeaders = {headers: {"Metadata-Flavor": "Google"}}
const instanceName = await (await fetch("http://169.254.169.254/computeMetadata/v1/instance/name", metadataHeaders)).text()
const zone = await (await fetch("http://169.254.169.254/computeMetadata/v1/instance/zone", metadataHeaders)).text()
const projectId = await (await fetch("http://169.254.169.254/computeMetadata/v1/project/project-id", metadataHeaders)).text()
const token = await (await fetch("http://169.254.169.254/computeMetadata/v1/instance/service-accounts/default/token", metadataHeaders)).json()

const sp = zone.split("/")
const urlPath = [sp[0], projectId, ...sp.slice(2)].join("/")

// debug(`token: ${JSON.stringify(token)}`)
debug(`accessToken: ${token["access_token"]}, zone: ${urlPath}, instanceName: ${instanceName}`)

while (true) {
  const out = await (await fetch(`https://compute.googleapis.com/compute/v1/${urlPath}/instances/${instanceName}`, {headers: {"Accept": "application/json", "Authorization": `Bearer ${token["access_token"]}`}})).json()

  // debug(`out: ${JSON.stringify(out)}`)
  debug(`current status: ${out["status"]}`)

  if (out["status"] != "PROVISIONING" && out["status"] != "STAGING" && out["status"] != "RUNNING") {
    info(`instance has status (${out["status"]}), marking it for deletion`)
    info("deleting node")
    await $`kubectl delete --wait=false nodes.clusters.kloudlite.io/${nodename}`
    await $`kubectl drain --ignore-daemonsets --delete-emptydir-data --force ${nodename} --timeout 10s`.nothrow()
    // await $`kubectl drain --ignore-daemonsets --delete-emptydir-data --force ${nodename} --timeout 10s`.nothrow()
    // await $`kubectl delete node/${nodename}`
    process.exit(0)
  }
  await sleep(1000);
}
