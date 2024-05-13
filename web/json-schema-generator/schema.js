// Generated file, Generated for validators

const schema = {
  "ConsoleUpdateAppMutationVariables": {
    "type": "object",
    "properties": {
      "envName": {
        "type": "string"
      },
      "app": {
        "type": "object",
        "properties": {
          "apiVersion": {
            "type": "string"
          },
          "ciBuildId": {
            "type": "string"
          },
          "displayName": {
            "type": "string"
          },
          "enabled": {
            "type": "boolean"
          },
          "kind": {
            "type": "string"
          },
          "metadata": {
            "type": "object",
            "properties": {
              "annotations": {},
              "labels": {},
              "name": {
                "type": "string"
              },
              "namespace": {
                "type": "string"
              }
            },
            "required": [
              "name"
            ]
          },
          "spec": {
            "type": "object",
            "properties": {
              "containers": {
                "type": "array",
                "items": {
                  "type": "object",
                  "properties": {
                    "args": {
                      "type": "array",
                      "items": {
                        "type": "string"
                      }
                    },
                    "command": {
                      "type": "array",
                      "items": {
                        "type": "string"
                      }
                    },
                    "env": {
                      "type": "array",
                      "items": {
                        "type": "object",
                        "properties": {
                          "key": {
                            "type": "string"
                          },
                          "optional": {
                            "type": "boolean"
                          },
                          "refKey": {
                            "type": "string"
                          },
                          "refName": {
                            "type": "string"
                          },
                          "type": {
                            "$ref": "#/definitions/InputMaybe"
                          },
                          "value": {
                            "type": "string"
                          }
                        },
                        "required": [
                          "key"
                        ]
                      }
                    },
                    "envFrom": {
                      "type": "array",
                      "items": {
                        "type": "object",
                        "properties": {
                          "refName": {
                            "type": "string"
                          },
                          "type": {
                            "$ref": "#/definitions/Github__Com___Kloudlite___Operator___Apis___Crds___V1__ConfigOrSecret"
                          }
                        },
                        "required": [
                          "refName",
                          "type"
                        ]
                      }
                    },
                    "image": {
                      "type": "string"
                    },
                    "imagePullPolicy": {
                      "type": "string"
                    },
                    "livenessProbe": {
                      "type": "object",
                      "properties": {
                        "failureThreshold": {
                          "type": "number"
                        },
                        "httpGet": {
                          "type": "object",
                          "properties": {
                            "httpHeaders": {},
                            "path": {
                              "type": "string"
                            },
                            "port": {
                              "type": "number"
                            }
                          },
                          "required": [
                            "path",
                            "port"
                          ]
                        },
                        "initialDelay": {
                          "type": "number"
                        },
                        "interval": {
                          "type": "number"
                        },
                        "shell": {
                          "type": "object",
                          "properties": {
                            "command": {
                              "type": "array",
                              "items": {
                                "type": "string"
                              }
                            }
                          }
                        },
                        "tcp": {
                          "type": "object",
                          "properties": {
                            "port": {
                              "type": "number"
                            }
                          },
                          "required": [
                            "port"
                          ]
                        },
                        "type": {
                          "type": "string"
                        }
                      },
                      "required": [
                        "type"
                      ]
                    },
                    "name": {
                      "type": "string"
                    },
                    "readinessProbe": {
                      "type": "object",
                      "properties": {
                        "failureThreshold": {
                          "type": "number"
                        },
                        "httpGet": {
                          "type": "object",
                          "properties": {
                            "httpHeaders": {},
                            "path": {
                              "type": "string"
                            },
                            "port": {
                              "type": "number"
                            }
                          },
                          "required": [
                            "path",
                            "port"
                          ]
                        },
                        "initialDelay": {
                          "type": "number"
                        },
                        "interval": {
                          "type": "number"
                        },
                        "shell": {
                          "type": "object",
                          "properties": {
                            "command": {
                              "type": "array",
                              "items": {
                                "type": "string"
                              }
                            }
                          }
                        },
                        "tcp": {
                          "type": "object",
                          "properties": {
                            "port": {
                              "type": "number"
                            }
                          },
                          "required": [
                            "port"
                          ]
                        },
                        "type": {
                          "type": "string"
                        }
                      },
                      "required": [
                        "type"
                      ]
                    },
                    "resourceCpu": {
                      "type": "object",
                      "properties": {
                        "max": {
                          "type": "string"
                        },
                        "min": {
                          "type": "string"
                        }
                      }
                    },
                    "resourceMemory": {
                      "type": "object",
                      "properties": {
                        "max": {
                          "type": "string"
                        },
                        "min": {
                          "type": "string"
                        }
                      }
                    },
                    "volumes": {
                      "type": "array",
                      "items": {
                        "type": "object",
                        "properties": {
                          "items": {
                            "type": "array",
                            "items": {
                              "type": "object",
                              "properties": {
                                "fileName": {
                                  "type": "string"
                                },
                                "key": {
                                  "type": "string"
                                }
                              },
                              "required": [
                                "key"
                              ]
                            }
                          },
                          "mountPath": {
                            "type": "string"
                          },
                          "refName": {
                            "type": "string"
                          },
                          "type": {
                            "$ref": "#/definitions/Github__Com___Kloudlite___Operator___Apis___Crds___V1__ConfigOrSecret"
                          }
                        },
                        "required": [
                          "mountPath",
                          "refName",
                          "type"
                        ]
                      }
                    }
                  },
                  "required": [
                    "image",
                    "name"
                  ]
                }
              },
              "displayName": {
                "type": "string"
              },
              "freeze": {
                "type": "boolean"
              },
              "hpa": {
                "type": "object",
                "properties": {
                  "enabled": {
                    "type": "boolean"
                  },
                  "maxReplicas": {
                    "type": "number"
                  },
                  "minReplicas": {
                    "type": "number"
                  },
                  "thresholdCpu": {
                    "type": "number"
                  },
                  "thresholdMemory": {
                    "type": "number"
                  }
                },
                "required": [
                  "enabled"
                ]
              },
              "intercept": {
                "type": "object",
                "properties": {
                  "enabled": {
                    "type": "boolean"
                  },
                  "portMappings": {
                    "type": "array",
                    "items": {
                      "type": "object",
                      "properties": {
                        "appPort": {
                          "type": "number"
                        },
                        "devicePort": {
                          "type": "number"
                        }
                      },
                      "required": [
                        "appPort",
                        "devicePort"
                      ]
                    }
                  },
                  "toDevice": {
                    "type": "string"
                  }
                },
                "required": [
                  "enabled",
                  "toDevice"
                ]
              },
              "nodeSelector": {},
              "region": {
                "type": "string"
              },
              "replicas": {
                "type": "number"
              },
              "router": {
                "type": "object",
                "properties": {
                  "backendProtocol": {
                    "type": "string"
                  },
                  "basicAuth": {
                    "type": "object",
                    "properties": {
                      "enabled": {
                        "type": "boolean"
                      },
                      "secretName": {
                        "type": "string"
                      },
                      "username": {
                        "type": "string"
                      }
                    },
                    "required": [
                      "enabled"
                    ]
                  },
                  "cors": {
                    "type": "object",
                    "properties": {
                      "allowCredentials": {
                        "type": "boolean"
                      },
                      "enabled": {
                        "type": "boolean"
                      },
                      "origins": {
                        "type": "array",
                        "items": {
                          "type": "string"
                        }
                      }
                    }
                  },
                  "domains": {
                    "type": "array",
                    "items": {
                      "type": "string"
                    }
                  },
                  "https": {
                    "type": "object",
                    "properties": {
                      "clusterIssuer": {
                        "type": "string"
                      },
                      "enabled": {
                        "type": "boolean"
                      },
                      "forceRedirect": {
                        "type": "boolean"
                      }
                    },
                    "required": [
                      "enabled"
                    ]
                  },
                  "ingressClass": {
                    "type": "string"
                  },
                  "maxBodySizeInMB": {
                    "type": "number"
                  },
                  "rateLimit": {
                    "type": "object",
                    "properties": {
                      "connections": {
                        "type": "number"
                      },
                      "enabled": {
                        "type": "boolean"
                      },
                      "rpm": {
                        "type": "number"
                      },
                      "rps": {
                        "type": "number"
                      }
                    }
                  },
                  "routes": {
                    "type": "array",
                    "items": {
                      "type": "object",
                      "properties": {
                        "app": {
                          "type": "string"
                        },
                        "path": {
                          "type": "string"
                        },
                        "port": {
                          "type": "number"
                        },
                        "rewrite": {
                          "type": "boolean"
                        }
                      },
                      "required": [
                        "app",
                        "path",
                        "port"
                      ]
                    }
                  }
                },
                "required": [
                  "domains"
                ]
              },
              "serviceAccount": {
                "type": "string"
              },
              "services": {
                "type": "array",
                "items": {
                  "type": "object",
                  "properties": {
                    "port": {
                      "type": "number"
                    },
                    "protocol": {
                      "type": "string"
                    }
                  },
                  "required": [
                    "port"
                  ]
                }
              },
              "tolerations": {
                "type": "array",
                "items": {
                  "type": "object",
                  "properties": {
                    "effect": {
                      "$ref": "#/definitions/InputMaybe_1"
                    },
                    "key": {
                      "type": "string"
                    },
                    "operator": {
                      "$ref": "#/definitions/InputMaybe_2"
                    },
                    "tolerationSeconds": {
                      "type": "number"
                    },
                    "value": {
                      "type": "string"
                    }
                  }
                }
              },
              "topologySpreadConstraints": {
                "type": "array",
                "items": {
                  "type": "object",
                  "properties": {
                    "labelSelector": {
                      "type": "object",
                      "properties": {
                        "matchExpressions": {
                          "type": "array",
                          "items": {
                            "type": "object",
                            "properties": {
                              "key": {
                                "type": "string"
                              },
                              "operator": {
                                "$ref": "#/definitions/K8s__Io___Apimachinery___Pkg___Apis___Meta___V1__LabelSelectorOperator"
                              },
                              "values": {
                                "type": "array",
                                "items": {
                                  "type": "string"
                                }
                              }
                            },
                            "required": [
                              "key",
                              "operator"
                            ]
                          }
                        },
                        "matchLabels": {}
                      }
                    },
                    "matchLabelKeys": {
                      "type": "array",
                      "items": {
                        "type": "string"
                      }
                    },
                    "maxSkew": {
                      "type": "number"
                    },
                    "minDomains": {
                      "type": "number"
                    },
                    "nodeAffinityPolicy": {
                      "type": "string"
                    },
                    "nodeTaintsPolicy": {
                      "type": "string"
                    },
                    "topologyKey": {
                      "type": "string"
                    },
                    "whenUnsatisfiable": {
                      "$ref": "#/definitions/K8s__Io___Api___Core___V1__UnsatisfiableConstraintAction"
                    }
                  },
                  "required": [
                    "maxSkew",
                    "topologyKey",
                    "whenUnsatisfiable"
                  ]
                }
              }
            },
            "required": [
              "containers"
            ]
          }
        },
        "required": [
          "displayName",
          "spec"
        ]
      }
    },
    "required": [
      "app",
      "envName"
    ],
    "definitions": {
      "InputMaybe": {
        "enum": [
          "config",
          "pvc",
          "secret"
        ],
        "type": "string"
      },
      "Github__Com___Kloudlite___Operator___Apis___Crds___V1__ConfigOrSecret": {
        "enum": [
          "config",
          "pvc",
          "secret"
        ],
        "type": "string"
      },
      "InputMaybe_1": {
        "enum": [
          "NoExecute",
          "NoSchedule",
          "PreferNoSchedule"
        ],
        "type": "string"
      },
      "InputMaybe_2": {
        "enum": [
          "Equal",
          "Exists"
        ],
        "type": "string"
      },
      "K8s__Io___Apimachinery___Pkg___Apis___Meta___V1__LabelSelectorOperator": {
        "enum": [
          "DoesNotExist",
          "Exists",
          "In",
          "NotIn"
        ],
        "type": "string"
      },
      "K8s__Io___Api___Core___V1__UnsatisfiableConstraintAction": {
        "enum": [
          "DoNotSchedule",
          "ScheduleAnyway"
        ],
        "type": "string"
      }
    },
    "$schema": "http://json-schema.org/draft-07/schema#"
  }
};

export default schema;