import { Map } from "immutable";

export interface CronTask {
  func: (arg?: any, ...optionalArgs: any[]) => void;
  delay: number;
  args?: any[];
  handler?: number;
}

export class Cron {
  private tasks: Map<string, CronTask>;
  constructor() {
    this.tasks = Map<string, CronTask>();
  }

  setInterval = (name: string, task: CronTask) => {
    if (this.tasks.has(name)) {
      this.clearInterval(name);
    }

    const handler = window.setInterval(task.func, task.delay, ...task.args);
    task.handler = handler;
    this.tasks = this.tasks.set(name, task);
  };

  clearInterval = (name: string) => {
    const preTask = this.tasks.get(name);
    window.clearInterval(preTask.handler);
    this.tasks = this.tasks.delete(name);
  };

  getTasks = (): Map<string, CronTask> => {
    return this.tasks;
  };
}

const cronJobs = new Cron();
export const CronJobs = (): Cron => {
  return cronJobs;
};
