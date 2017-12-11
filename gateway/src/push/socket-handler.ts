"use strict";
import * as SocketIO from "socket.io";
import * as log from "winston";
import { SocketConnection } from "./socket-connection";

export class SocketHandler {
  private static isInitialized: boolean = false;
  private static connectionsMap: { [id: string]: SocketConnection[]} = {}; // map of namespaces' connections

  public static initialize() {
    SocketHandler.isInitialized = true;
  }

  public static onConnection(namespace: string, socket: SocketIO.Socket) {
    if (!SocketHandler.isInitialized) {
      log.info("Closing connection to websocket (handler not yet initialized)");
      socket.disconnect(true);
      return;
    }

    if (namespace === "occupancy") {

    } else {
      log.warn("Closing connection to websocket with unknown namespace");
      socket.disconnect(true);
    }
  }

  private static initConnection(namespace: string, socket: SocketIO.Socket, connection: SocketConnection): void {
    socket.on("disconnect", () => this.onDisconnect(namespace, connection));
    SocketHandler.connectionsMap[namespace].push(connection);
  }

  private static onDisconnect(namespace: string, connection: SocketConnection): void {
    const connections = this.connectionsMap[namespace];

    for (let i = 0; i < connections.length; i++) {
      if (connections[i].getConnectionId() === connection.getConnectionId()) {
        connections.splice(i, 1);
        log.info("Socket successfully disconnected");
        return;
      }
    }

    log.warn("Got disconnect from unknown socket");
  }

  public static getConnection(namespace: string): SocketConnection[] {
    return this.connectionsMap[namespace];
  }
}