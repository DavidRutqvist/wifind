import { Model } from "mongoose";
import { ISensorLocationModel } from "./sensor-location";

export interface IModel {
    sensorLocation: Model<ISensorLocationModel>;
}