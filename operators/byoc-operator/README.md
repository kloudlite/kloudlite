[BYOC] 
    -> registerd in database
    -> kloudlite cluster, but missing critical defaults

    {Agent}
        -> dispatches defaults for BYOC like
            + default storage class 
            + default ingress class
            + public IPs

    [Message Consumer]
       -> reads from queue, to patch default values to resource spec
       -> updates them in DB
