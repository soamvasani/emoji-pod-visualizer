from flask import request
from flask import current_app
import json
import requests

def main():
    eventType = request.headers["X-Kubernetes-Event-Type"]
    objType = request.headers["X-Kubernetes-Object-Type"]

    if objType != "Pod":
        return
    
    pod = request.get_json()
    if (pod == None) or ("status" not in pod) or ("containerStatuses" not in pod["status"]):
        current_app.logger.info("No container status information in event")
        return ""

    podName = pod["metadata"]["name"]
    current_app.logger.info("Event %s for %s %s" % (eventType, objType, podName))

    containerStatuses = pod["status"]["containerStatuses"]

    states = {}
    for cs in containerStatuses:
        name = cs["name"]
        state = cs["state"]
        if "running" in state:
            states[name] = "running"
        elif "terminated" in state:
            states[name] = "terminated"
        else:
            s = "waiting"
            if ("waiting" in state) and ("reason" in state["waiting"]):
                s = state["waiting"]["reason"]
            states[name] = s

    current_app.logger.info("C states = %s" % states)
    
    # Notify clients
    notification = { "numPods": 1, "containerStates": states, "podName": podName, "eventType": eventType }
    r = requests.post("http://podvis.default:8001/", data=json.dumps(notification))
    current_app.logger.info("Pod vis event post status: %s" % r.status_code)
    
    return ""
