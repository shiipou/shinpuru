/** @format */

export interface ListReponse<T> {
  n: number;
  data: T[];
}

export interface User {
  id: string;
  username: string;
  avatar: string;
  locale: string;
  discriminator: string;
  verified: boolean;
  bot: boolean;
  avatar_url: string;
}

export interface Role {
  id: string;
  name: string;
  managed: boolean;
  mentionable: boolean;
  hoist: boolean;
  color: number;
  position: number;
  permission: number;
}

export interface Member {
  user: User;
  guild_id: string;
  joined_at: string;
  nick: string;
  roles: string[];
}

export interface Channel {
  id: string;
  guild_id: string;
  name: string;
  topic: string;
  type: number;
  nsfw: boolean;
  position: number;
  user_limit: number;
  parent_id: string;
}

export interface Guild {
  id: string;
  name: string;
  icon: string;
  icon_url: string;
  region: string;
  owner_id: string;
  joined_at: string;
  member_count: number;

  roles?: Role[];
  members?: Member[];
  channels?: Channel[];
}