import { NextRequest, NextResponse } from 'next/server'

export async function GET(request: NextRequest) {
  const searchParams = request.nextUrl.searchParams
  const key = searchParams.get('key')
  const location = searchParams.get('location')

  if (!key) {
    return NextResponse.json({ error: 'Missing key parameter' }, { status: 400 })
  }

  if (!location) {
    return NextResponse.json({ error: 'Missing location parameter' }, { status: 400 })
  }

  const template = {
    "$schema": "https://schema.management.azure.com/schemas/2018-05-01/subscriptionDeploymentTemplate.json#",
    "contentVersion": "1.0.0.0",
    "metadata": {
      "title": "Kloudlite One-Click Installation",
      "description": "Deploy Kloudlite development workspace platform to your Azure subscription"
    },
    "parameters": {},
    "variables": {
      "location": location,
      "kloudliteKey": key,
      "resourceGroupName": `[concat('kloudlite-installer-', uniqueString(subscription().subscriptionId, '${location}'))]`,
      "vmName": "[concat('kl-installer-', uniqueString(subscription().subscriptionId))]",
      "identityName": "[concat('kl-installer-id-', uniqueString(subscription().subscriptionId))]",
      "nicName": "[concat('kl-installer-', uniqueString(subscription().subscriptionId), '-nic')]",
      "nsgName": "[concat('kl-installer-', uniqueString(subscription().subscriptionId), '-nsg')]",
      "vnetName": "[concat('kl-installer-', uniqueString(subscription().subscriptionId), '-vnet')]",
      "pipName": "[concat('kl-installer-', uniqueString(subscription().subscriptionId), '-pip')]"
    },
    "resources": [
      {
        "type": "Microsoft.Resources/resourceGroups",
        "apiVersion": "2022-09-01",
        "name": "[variables('resourceGroupName')]",
        "location": "[variables('location')]",
        "tags": {
          "kloudlite-managed-by": "arm-template",
          "kloudlite-purpose": "installer"
        }
      },
      {
        "type": "Microsoft.Resources/deployments",
        "apiVersion": "2022-09-01",
        "name": "identityDeployment",
        "resourceGroup": "[variables('resourceGroupName')]",
        "dependsOn": [
          "[resourceId('Microsoft.Resources/resourceGroups', variables('resourceGroupName'))]"
        ],
        "properties": {
          "mode": "Incremental",
          "expressionEvaluationOptions": {
            "scope": "inner"
          },
          "template": {
            "$schema": "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#",
            "contentVersion": "1.0.0.0",
            "parameters": {
              "identityNameParam": { "type": "string" },
              "locationParam": { "type": "string" }
            },
            "variables": {
              "identityName": "[parameters('identityNameParam')]",
              "location": "[parameters('locationParam')]"
            },
            "resources": [
              {
                "type": "Microsoft.ManagedIdentity/userAssignedIdentities",
                "apiVersion": "2023-01-31",
                "name": "[variables('identityName')]",
                "location": "[variables('location')]"
              }
            ],
            "outputs": {
              "principalId": {
                "type": "string",
                "value": "[reference(variables('identityName'), '2023-01-31').principalId]"
              },
              "identityId": {
                "type": "string",
                "value": "[resourceId('Microsoft.ManagedIdentity/userAssignedIdentities', variables('identityName'))]"
              }
            }
          },
          "parameters": {
            "identityNameParam": { "value": "[variables('identityName')]" },
            "locationParam": { "value": "[variables('location')]" }
          }
        }
      },
      {
        "type": "Microsoft.Resources/deployments",
        "apiVersion": "2022-09-01",
        "name": "vmDeployment",
        "resourceGroup": "[variables('resourceGroupName')]",
        "dependsOn": [
          "identityDeployment"
        ],
        "properties": {
          "mode": "Incremental",
          "expressionEvaluationOptions": {
            "scope": "inner"
          },
          "template": {
            "$schema": "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#",
            "contentVersion": "1.0.0.0",
            "parameters": {
              "vmName": { "type": "string" },
              "identityId": { "type": "string" },
              "principalId": { "type": "string" },
              "location": { "type": "string" },
              "resourceGroupName": { "type": "string" },
              "roleAssignmentGuid": { "type": "string" },
              "kloudliteKey": { "type": "string" },
              "nicName": { "type": "string" },
              "nsgName": { "type": "string" },
              "vnetName": { "type": "string" },
              "pipName": { "type": "string" }
            },
            "resources": [
              {
                "type": "Microsoft.Authorization/roleAssignments",
                "apiVersion": "2022-04-01",
                "name": "[parameters('roleAssignmentGuid')]",
                "properties": {
                  "roleDefinitionId": "[subscriptionResourceId('Microsoft.Authorization/roleDefinitions', 'b24988ac-6180-42a0-ab88-20f7382dd24c')]",
                  "principalId": "[parameters('principalId')]",
                  "principalType": "ServicePrincipal"
                }
              },
              {
                "type": "Microsoft.Network/networkSecurityGroups",
                "apiVersion": "2023-05-01",
                "name": "[parameters('nsgName')]",
                "location": "[parameters('location')]",
                "properties": {
                  "securityRules": []
                }
              },
              {
                "type": "Microsoft.Network/publicIPAddresses",
                "apiVersion": "2023-05-01",
                "name": "[parameters('pipName')]",
                "location": "[parameters('location')]",
                "sku": { "name": "Basic" },
                "properties": { "publicIPAllocationMethod": "Dynamic" }
              },
              {
                "type": "Microsoft.Network/virtualNetworks",
                "apiVersion": "2023-05-01",
                "name": "[parameters('vnetName')]",
                "location": "[parameters('location')]",
                "dependsOn": ["[resourceId('Microsoft.Network/networkSecurityGroups', parameters('nsgName'))]"],
                "properties": {
                  "addressSpace": { "addressPrefixes": ["10.0.0.0/16"] },
                  "subnets": [{
                    "name": "default",
                    "properties": {
                      "addressPrefix": "10.0.0.0/24",
                      "networkSecurityGroup": { "id": "[resourceId('Microsoft.Network/networkSecurityGroups', parameters('nsgName'))]" }
                    }
                  }]
                }
              },
              {
                "type": "Microsoft.Network/networkInterfaces",
                "apiVersion": "2023-05-01",
                "name": "[parameters('nicName')]",
                "location": "[parameters('location')]",
                "dependsOn": [
                  "[resourceId('Microsoft.Network/virtualNetworks', parameters('vnetName'))]",
                  "[resourceId('Microsoft.Network/publicIPAddresses', parameters('pipName'))]"
                ],
                "properties": {
                  "ipConfigurations": [{
                    "name": "ipconfig1",
                    "properties": {
                      "privateIPAllocationMethod": "Dynamic",
                      "subnet": { "id": "[resourceId('Microsoft.Network/virtualNetworks/subnets', parameters('vnetName'), 'default')]" },
                      "publicIPAddress": { "id": "[resourceId('Microsoft.Network/publicIPAddresses', parameters('pipName'))]" }
                    }
                  }]
                }
              },
              {
                "type": "Microsoft.Compute/virtualMachines",
                "apiVersion": "2023-07-01",
                "name": "[parameters('vmName')]",
                "location": "[parameters('location')]",
                "dependsOn": [
                  "[resourceId('Microsoft.Network/networkInterfaces', parameters('nicName'))]",
                  "[resourceId('Microsoft.Authorization/roleAssignments', parameters('roleAssignmentGuid'))]"
                ],
                "identity": {
                  "type": "UserAssigned",
                  "userAssignedIdentities": {
                    "[parameters('identityId')]": {}
                  }
                },
                "properties": {
                  "hardwareProfile": { "vmSize": "Standard_B1s" },
                  "osProfile": {
                    "computerName": "[parameters('vmName')]",
                    "adminUsername": "azureuser",
                    "adminPassword": "[concat('P@ss', uniqueString(parameters('resourceGroupName')), '!')]",
                    "linuxConfiguration": { "disablePasswordAuthentication": false }
                  },
                  "storageProfile": {
                    "imageReference": {
                      "publisher": "Canonical",
                      "offer": "0001-com-ubuntu-server-jammy",
                      "sku": "22_04-lts-gen2",
                      "version": "latest"
                    },
                    "osDisk": {
                      "createOption": "FromImage",
                      "managedDisk": { "storageAccountType": "Standard_LRS" }
                    }
                  },
                  "networkProfile": {
                    "networkInterfaces": [{ "id": "[resourceId('Microsoft.Network/networkInterfaces', parameters('nicName'))]" }]
                  }
                }
              },
              {
                "type": "Microsoft.Compute/virtualMachines/extensions",
                "apiVersion": "2023-07-01",
                "name": "[concat(parameters('vmName'), '/InstallKloudlite')]",
                "location": "[parameters('location')]",
                "dependsOn": [
                  "[resourceId('Microsoft.Compute/virtualMachines', parameters('vmName'))]"
                ],
                "properties": {
                  "publisher": "Microsoft.Azure.Extensions",
                  "type": "CustomScript",
                  "typeHandlerVersion": "2.1",
                  "autoUpgradeMinorVersion": true,
                  "settings": {
                    "script": "[base64(concat('#!/bin/bash\nset -euo pipefail\nexec > >(tee /var/log/kloudlite-install.log) 2>&1\necho \"Starting Kloudlite installation...\"\ncurl -fsSL https://get.khost.dev/install/azure | bash -s -- --key \"', parameters('kloudliteKey'), '\" --location \"', parameters('location'), '\"\necho \"Installation complete. Cleaning up installer in 60 seconds...\"\nsleep 60\necho \"Installing Azure CLI for cleanup...\"\ncurl -sL https://aka.ms/InstallAzureCLIDeb | sudo bash\necho \"Logging in with managed identity...\"\naz login --identity --allow-no-subscriptions\necho \"Deleting installer resource group...\"\naz group delete --name ', parameters('resourceGroupName'), ' --yes --no-wait\necho \"Cleanup initiated.\"'))]"
                  }
                }
              }
            ]
          },
          "parameters": {
            "vmName": { "value": "[variables('vmName')]" },
            "identityId": { "value": "[reference('identityDeployment').outputs.identityId.value]" },
            "principalId": { "value": "[reference('identityDeployment').outputs.principalId.value]" },
            "location": { "value": "[variables('location')]" },
            "resourceGroupName": { "value": "[variables('resourceGroupName')]" },
            "roleAssignmentGuid": { "value": "[guid(variables('resourceGroupName'), reference('identityDeployment').outputs.principalId.value, 'Contributor')]" },
            "kloudliteKey": { "value": "[variables('kloudliteKey')]" },
            "nicName": { "value": "[variables('nicName')]" },
            "nsgName": { "value": "[variables('nsgName')]" },
            "vnetName": { "value": "[variables('vnetName')]" },
            "pipName": { "value": "[variables('pipName')]" }
          }
        }
      }
    ],
    "outputs": {
      "resourceGroupName": {
        "type": "string",
        "value": "[variables('resourceGroupName')]"
      },
      "location": {
        "type": "string",
        "value": "[variables('location')]"
      }
    }
  }

  return NextResponse.json(template, {
    headers: {
      'Content-Type': 'application/json',
      'Access-Control-Allow-Origin': '*',
      'Cache-Control': 'no-cache, no-store, must-revalidate'
    }
  })
}
