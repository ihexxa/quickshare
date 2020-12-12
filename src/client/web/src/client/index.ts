import { FilesClient } from "../client/files";
import { UsersClient } from "../client/users";

export const filesClient = new FilesClient("http://127.0.0.1:8888");
export const usersClient = new UsersClient("http://127.0.0.1:8888");