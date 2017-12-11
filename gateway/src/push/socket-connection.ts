"use strict";
import * as SocketIO from "socket.io";
import * as uuid from "uuid/v4";

export abstract class SocketConnection {
  private readonly connectionId: string;
  
  constructor(private readonly namespace: string, private readonly socket: SocketIO.Socket) {
    this.connectionId = uuid();
  }
  
  public push(event: string, data?: any): void {
    if (data) {
      this.socket.emit(event, JSON.stringify(data));
    } else {
      this.socket.emit(event);
    }
  }

  public getConnectionId(): string {
    return this.connectionId;
  }
}