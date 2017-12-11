"use strict";
import { SocketConnection } from "../socket-connection";

export class OccupancyConnection extends SocketConnection {
  private readonly zoneOccupancyCache: { [id: string]: number} = {};

  constructor(private readonly prenumerations: string[], socket: SocketIO.Socket) {
    super("occupancy", socket);
  }

  public occupancyUpdated(zoneId: string, occupancy: number): void {
    if (this.containsPrenumeration(zoneId)) {
      if ((this.zoneOccupancyCache[zoneId] === undefined) || (this.zoneOccupancyCache[zoneId] !== occupancy)) {
        this.zoneOccupancyCache[zoneId] = occupancy;
        this.push("OCCUPANCY_CHANGED", {
          zone: zoneId,
          occupancy: occupancy
        });
      }
    }
  }

  private containsPrenumeration(zoneId: string): boolean {
    for (let i = 0; i < this.prenumerations.length; i++) {
      if (this.prenumerations[i] === zoneId) {
        return true;
      }
    }

    return false;
  }
}
