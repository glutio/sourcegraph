/** Partial contribution-related client capabilities. */
export interface ContributionClientCapabilities {
    /** The window client capabilities. */
    window?: {
        /** Contribution-related client capabilities. */
        contribution?: {
            /** Whether the client supports dynamic registration of contributions. */
            dynamicRegistration?: boolean
        }
    }
}

/** Partial contribution-related server capabilities. */
export interface ContributionServerCapabilities {
    /** The contributions provided by the server. */
    contributions?: Contributions
}

/**
 * Contributions describes the functionality provided by an extension.
 */
export interface Contributions {
    /** Commands contributed by the extension. */
    commands?: CommandContribution[]

    /** Menu items contributed by the extension. */
    menus?: MenuContributions
}

/**
 * CommandContribution is a command provided by the extension that can be invoked.
 */
export interface CommandContribution {
    /**
     * Command is an identifier for the command that is assumed to be unique. If another command with the same
     * identifier is defined (by this extension or another extension), the behavior is undefined. To avoid
     * collisions, the identifier conventionally is prefixed with "${EXTENSION_NAME}.".
     */
    command: string

    /** A descriptive title. */
    title?: string

    /** A URL to an icon (base64: URIs are OK). */
    iconURL?: string

    /**
     * TODO: Because the CXP connection is (usually) stateless, commands can't modify state. The second best option
     * is for them to modify user settings. So, require commands to define how they do so.
     */
    experimentalSettingsAction?: CommandContributionSettingsAction
}

/**
 * CommandContributionSettingsAction is the special action executed by a contributed command that modifies
 * settings.
 */
interface CommandContributionSettingsAction {
    /** The key path to the value. */
    path: (string | number)[]

    // Exactly 1 of the following fields must be set.

    /** The values of the setting to cycle among. */
    cycleValues?: any[]
    /** Show a user prompt to obtain the value for the setting. */
    prompt?: string
}

export enum ContributableMenu {
    /** The global command palette. */
    CommandPalette = 'commandPalette',

    /** The global navigation bar in the application. */
    GlobalNav = 'global/nav',

    /** The title bar for the current document. */
    EditorTitle = 'editor/title',
}

/**
 * MenuContributions describes the menu items contributed by an extension.
 */
export interface MenuContributions extends Partial<Record<ContributableMenu, MenuItemContribution[]>> {}

/**
 * MenuItemContribution is a menu item contributed by an extension.
 */
export interface MenuItemContribution {
    /** The command to execute when selected (== (CommandContribution).command). */
    command: string
}