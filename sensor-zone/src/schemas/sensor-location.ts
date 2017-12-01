'use strict';
import { Schema } from "mongoose";

export var sensorLocationSchema: Schema = new Schema({
  sensorId: { type: String, index: true, required: true },
  zoneId: { type: Schema.Types.ObjectId, index: true, required: true },
  from: { type: Date, required: true },
  to: { type: Date, required: false }
});

sensorLocationSchema.index({ sensorId: 1, from: 1 });