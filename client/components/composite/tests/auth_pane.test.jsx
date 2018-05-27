import React from "react";
import { AuthPane, classLogin, classLogout } from "../auth_pane";

describe("AuthPane", () => {
  test("AuthPane should show login pane if isLogin === true, or show logout pane", () => {
    const tests = [
      {
        input: {
          onLogin: jest.fn,
          onLogout: jest.fn,
          isLogin: false,
          serverAddr: ""
        },
        output: classLogin
      },
      {
        input: {
          onLogin: jest.fn,
          onLogout: jest.fn,
          isLogin: true,
          serverAddr: ""
        },
        output: classLogout
      }
    ];

    tests.forEach(testCase => {
      const pane = new AuthPane(testCase.input);
      expect(pane.render().props.className).toBe(testCase.output);
    });
  });
});
