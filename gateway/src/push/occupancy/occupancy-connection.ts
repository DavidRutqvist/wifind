import { SocketConnection } from "../socket-connection";

"use strict";

export class OccupancyConnection extends SocketConnection {
  public occupancyUpdated(zoneId: string, occupancy: number): void {
    
  }
}