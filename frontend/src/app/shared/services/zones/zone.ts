"use strict";

export interface Zone {
  id: string;
  name: string;
  location: number[];
  parent?: string;
}
