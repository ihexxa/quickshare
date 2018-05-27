import React from "react";
import { mount } from "enzyme";
import { checkQueueCycle, Uploader } from "../uploader";

const testTimeout = 4000;

describe("Uploader", () => {
  test(
    "Uploader will upload files in uploadQueue by interval",
    () => {
      // TODO: could be refactored using timer mocks
      // https://facebook.github.io/jest/docs/en/timer-mocks.html
      const tests = [
        {
          input: { target: { files: ["task1", "task2", "task3"] } },
          uploadCalled: 3
        }
      ];

      let promises = [];

      const uploader = mount(<Uploader />);
      tests.forEach(testCase => {
        // mock
        const uploadSpy = jest.fn();
        const uploadStub = () => {
          uploadSpy();
          return Promise.resolve();
        };
        uploader.instance().upload = uploadStub;
        uploader.update();

        // upload and verify
        uploader.instance().onUpload(testCase.input);
        const wait = testCase.input.target.files.length * 1000 + 100;
        const promise = new Promise(resolve => {
          setTimeout(() => {
            expect(uploader.instance().state.uploadQueue.length).toBe(0);
            expect(uploadSpy.mock.calls.length).toBe(testCase.uploadCalled);
            resolve();
          }, wait);
        });

        promises = [...promises, promise];
      });

      return Promise.all(promises);
    },
    testTimeout
  );
});
