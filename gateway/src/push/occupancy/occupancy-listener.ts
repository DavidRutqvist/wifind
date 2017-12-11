"use strict";
import { AmqpTopicListener } from "../amqp-topic-listener";
import { Message } from "amqplib";
import { connect } from "net";

export class OccupancyListener extends AmqpTopicListener {
  constructor(connectionString: string, exchange: string) {
    super(connectionString, exchange, ["OCCUPANCY.#.UPDATED"]);
  }

  protected handle(message: Message): void {
    const routingParts = message.fields.routingKey ? message.fields.routingKey.split(".") : [];
    
    if (routingParts.length === 3) {
      const zoneId = routingParts[1];
      const data: any = JSON.parse(message.content.toString());
      
      
    }
  }
}