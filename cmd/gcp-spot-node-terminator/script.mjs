#! /usr/bin/env zx

$.verbose = false

const nodename = process.env.NODE_NAME
if (!nodename) {
  throw new Error("env var NODE_NAME is not set, exiting.")
}

const webhookURL = process.env.WEBHOOK_URL

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

async function getToken() {
  const t = await fetch("http://169.254.169.254/computeMetadata/v1/instance/service-accounts/default/token", metadataHeaders)
  const data = await t.json()
  return data["access_token"]
}

async function drainAndDeleteNode() {
  info("deleting node")
  await $`kubectl delete --wait=false nodes.clusters.kloudlite.io/${nodename}`.nothrow()
  await $`kubectl drain --ignore-daemonsets --delete-emptydir-data --force ${nodename} --timeout 10s`.nothrow()
}

const sp = zone.split("/")
const urlPath = [sp[0], projectId, ...sp.slice(2)].join("/")

let token = await getToken()

while (true) {
  const resp = await fetch(`https://compute.googleapis.com/compute/v1/${urlPath}/instances/${instanceName}`, {headers: {"Accept": "application/json", "Authorization": `Bearer ${token}`}})
  const out = await resp.json()

  if (out?.error?.code == 401) {
    // unauthorized, need to refresh token
    info("OAuth Token expired, refreshing token and retrying ...")
    token = await getToken()
    info("token refreshed")
    if (webhookURL) {
      await fetch(`https://${webhookURL}/push/text?message="OAuth Token Expired at (${nodename})"`)
    }
    continue
  }

  // debug(`out: ${JSON.stringify(out)}`)
  debug(`current status: ${out["status"]}`)
  if (out["status"] == "undefined") {
    info("google cloud instance get api is showing `undefined` as status of current node, will retry")
    await sleep(1000)
    continue
  }

  if (!["PROVISIONING", "STAGING", "RUNNING"].some(item => item == out["status"])) {
    if (webhookURL){
      await fetch(`${webhookURL}/push/text?message="proceeding with deletion when instance (${nodename}) status: ${out['status']}, \nfor debugging: full output\n${JSON.stringify(out)}"`)
    }
    info(`instance has status (${out["status"]}), marking it for deletion`)
    await drainAndDeleteNode()
    process.exit(0)
  }

  // checking preempted url
  const {stdout: preemptedStatus} = await $`curl "http://169.254.169.254/computeMetadata/v1/instance/preempted" -H "Metadata-Flavor: Google"`
  if (preemptedStatus.trim().toLowerCase() == "true") {
    info("node is preempted, exiting")
    if (webhookURL){
      await fetch(`${webhookURL}/push/text?message="Congratulations node (${nodename}) preempted by GCP"`)
    }
    await drainAndDeleteNode()
    process.exit(0)
  }

  await sleep(1000);
}
