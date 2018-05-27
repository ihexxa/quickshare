jest.mock("../../../libs/api_share");
jest.mock("../../../libs/api_auth");
import React from "react";
import { InfoBar } from "../info_bar";
import { mount } from "enzyme";
import * as mockApiShare from "../../../libs/api_share";
import { execFuncs, getDesc, verifyCalls } from "../../../tests/test_helper";
import valueEqual from "value-equal";

describe("InfoBar", () => {
  const tests = [
    {
      execs: [
        {
          func: "onSearch",
          args: ["searchFileName"]
        }
      ],
      state: {
        filterFileName: "searchFileName",
        fold: false
      },
      calls: [
        {
          func: "onSearch",
          count: 1
        }
      ]
    },
    {
      execs: [
        {
          func: "onLogin",
          args: ["serverAddr", "adminId", "adminPwd"]
        }
      ],
      state: {
        filterFileName: "",
        fold: false
      },
      calls: [
        {
          func: "onLogin",
          count: 1
        }
      ]
    },
    {
      execs: [
        {
          func: "onLogout",
          args: ["serverAddr"]
        }
      ],
      state: {
        filterFileName: "",
        fold: false
      },
      calls: [
        {
          func: "onLogout",
          count: 1
        }
      ]
    },
    {
      execs: [
        {
          func: "onAddLocalFiles",
          args: []
        }
      ],
      state: {
        filterFileName: "",
        fold: false
      },
      calls: [
        {
          func: "onAddLocalFiles",
          count: 1
        }
      ]
    }
  ];

  tests.forEach(testCase => {
    test(getDesc("InfoBar", testCase), () => {
      const stubs = {
        onLogin: jest.fn(),
        onLogout: jest.fn(),
        onAddLocalFiles: jest.fn(),
        onSearch: jest.fn(),
        onOk: jest.fn(),
        onError: jest.fn()
      };

      const onAddLocalFilesWrap = () => {
        stubs.onAddLocalFiles();
        return Promise.resolve(true);
      };

      return new Promise((resolve, reject) => {
        const infoBar = mount(
          <InfoBar
            width="100%"
            isLogin={false}
            serverAddr=""
            onLogin={stubs.onLogin}
            onLogout={stubs.onLogout}
            onAddLocalFiles={onAddLocalFilesWrap}
            onSearch={stubs.onSearch}
            onOk={stubs.onOk}
            onError={stubs.onError}
          />
        );

        execFuncs(infoBar.instance(), testCase.execs)
          .then(() => {
            infoBar.update();

            if (!valueEqual(infoBar.instance().state, testCase.state)) {
              return reject("state not identical");
            }
            if (testCase.calls != null) {
              const err = verifyCalls(testCase.calls, stubs);
              if (err !== null) {
                return reject(err);
              }
            }
            resolve();
          })
          .catch(err => console.error(err));
      });
    });
  });
});
