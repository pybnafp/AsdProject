import {randString} from "@/utils/libs";

export function getSessionId() {
    return randString(42)
}
