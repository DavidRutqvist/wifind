"use strict";
import { NextFunction, Request, Response } from "express";

export class AxiosHelper {
  public static handleError(error: any, res: Response): void {
    if(error.response && (error.response.status === 401)) {
      res.status(401).json({
        success: false,
        message: error.response.data.message
      });
    } else if(error.response && (error.response.status === 404)) {
      res.status(404).json({
        success: false,
        message: error.response.data.message
      });
    } else if(error.response && (error.response.status === 400) && (error.response.data.message)) {
      // all (most) checks for bad request, i.e. missing parameters and such should be explicit in gateway API
      // to not expose any internal architecture and maybe have more friendly error messages. However, we do allow
      // pass through mostly during development, we therefore log (warn) this before sending response.
      // There may be cases where it is inappropiate to make checks in Gateway as well, for example when trying to add
      // something and it already exists, this would require an extra call for a rare case. Consider using separate code
      // for this case, preferably 409.
      console.warn("Passed through 400 BAD REQUEST response, consider handling this explicitly in Gateway API " +
        "to not expose any internal structure. Response: " +
        error.response.data.message);

      res.status(400).json({
        success: false,
        message: error.response.data.message
      });
    } else {
      console.error(error);
      res.status(500).json({
        success: false,
        message: "Something went wrong internally."
      });
    }
  }
}
