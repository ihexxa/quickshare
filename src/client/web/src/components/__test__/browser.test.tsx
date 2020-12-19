import * as React from "react";

import { init } from "../core_state";
import { Updater } from "../auth_pane";
import { MockUsersClient } from "../../client/users_mock";
import { Response } from "../../client";

describe("Browser", () => {
  const makePromise = (ret: any): Promise<any> => {
    return new Promise<any>((resolve) => {
      resolve(ret);
    });
  };
  const makeNumberResponse = (status: number): Promise<Response> => {
    return makePromise({
      status: status,
      statusText: "",
      data: {},
    });
  };

  test("Updater:", async () => {
  });
});
